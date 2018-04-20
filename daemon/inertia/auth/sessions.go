package auth

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// session are properties associated with session,
// used for database entries
type session struct {
	Username string    `json:"username"`
	Expires  time.Time `json:"created"`
}

type sessionManager struct {
	cookieName    string
	cookieDomain  string
	cookieTimeout time.Duration
	internal      map[string]*session

	sync.RWMutex
	endSessionCleanup chan bool
}

func newSessionManager(domain string, timeout int) *sessionManager {
	manager := &sessionManager{
		cookieName:    "ubclaunchpad-inertia",
		cookieDomain:  domain,
		cookieTimeout: time.Duration(timeout) * time.Minute,
		internal:      make(map[string]*session),

		endSessionCleanup: make(chan bool),
	}

	// Set up session cleanup goroutine
	ticker := time.NewTicker(manager.cookieTimeout)
	go func() {
		for {
			select {
			case <-manager.endSessionCleanup:
				ticker.Stop()
				return
			case <-ticker.C:
				manager.Lock()
				for id, session := range manager.internal {
					if !manager.isValidSession(session) {
						delete(manager.internal, id)
					}
				}
				manager.Unlock()
			}
		}
	}()

	return manager
}

func (s *sessionManager) Close() {
	s.endSessionCleanup <- true

	s.Lock()
	s.internal = make(map[string]*session)
	s.Unlock()
}

// SessionBegin starts a new session with user by setting a cookie
// and adding session to memory
func (s *sessionManager) BeginSession(username string, w http.ResponseWriter, r *http.Request) error {
	expiration := time.Now().Add(s.cookieTimeout)
	id, err := generateSessionID()
	if err != nil {
		return errors.New("Failed to begin session for " + username + ": " + err.Error())
	}

	// Add session to map
	s.Lock()
	s.internal[id] = &session{
		Username: username,
		Expires:  expiration,
	}
	s.Unlock()

	// Add cookie with session ID
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    url.QueryEscape(id),
		Domain:   s.cookieDomain,
		MaxAge:   int(s.cookieTimeout / time.Second),
		HttpOnly: true,
		Expires:  expiration,
	})
	return nil
}

// SessionEnd ends a session and sets cookie to expire
func (s *sessionManager) EndSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.cookieName)
	if err != nil {
		return errors.New("Invalid cookie: " + err.Error())
	}
	if cookie.Value == "" {
		return errors.New("Invalid cookie")
	}
	id, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return errors.New("Invalid cookie: " + err.Error())
	}

	// Delete session from map
	s.Lock()
	delete(s.internal, id)
	s.Unlock()

	// Set cookie to expire immediately
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    url.QueryEscape(id),
		Domain:   s.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
	return nil
}

// GetSession verifies if given request is from a valid session and returns it
func (s *sessionManager) GetSession(w http.ResponseWriter, r *http.Request) (*session, error) {

	cookie, err := r.Cookie(s.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, errCookieNotFound
	}
	id, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}

	s.RLock()
	session, found := s.internal[id]
	if !found {
		s.RUnlock()
		return nil, errSessionNotFound
	}
	if !s.isValidSession(session) {
		s.RUnlock()
		s.EndSession(w, r)
		return nil, errSessionNotFound
	}
	s.RUnlock()

	return session, nil
}

// endAllUserSessions removes all active sessions with given user
func (s *sessionManager) EndAllUserSessions(username string) {
	s.Lock()
	for id, session := range s.internal {
		if session.Username == username {
			delete(s.internal, id)
		}
	}
	s.Unlock()
}

// EndAllSessions removes all active sessions
func (s *sessionManager) EndAllSessions() {
	s.Lock()
	s.internal = make(map[string]*session)
	s.Unlock()
}

// isValidSession checks if session is expired
func (s *sessionManager) isValidSession(session *session) bool {
	return session.Expires.After(time.Now())
}
