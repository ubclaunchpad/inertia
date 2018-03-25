package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/boltdb/bolt"
)

var (
	errSessionNotFound = errors.New("Session not found")
	errCookieNotFound  = errors.New("Cookie not found")
	errUserNotFound    = errors.New("User not found")
)

// userProps are properties associated with user, used
// for database entries
type userProps struct {
	HashedPassword string `json:"hashedPassword"`
	Admin          bool   `json:"admin"`
}

// session are properties associated with session,
// used for database entries
type session struct {
	Username string    `json:"username"`
	Expires  time.Time `json:"created"`
}

// userManager administers sessions and user accounts
type userManager struct {
	cookieName    string
	cookieDomain  string
	cookiePath    string
	cookieTimeout time.Duration

	// db is a boltdb database, which is an embedded
	// key/value database where each "bucket" is a collection
	db          *bolt.DB
	usersBucket []byte

	// sessions is a map of active user sessions
	sessions          map[string]*session
	endSessionCleanup chan bool
}

func newUserManager(dbPath, domain, path string, timeout int) (*userManager, error) {
	manager := &userManager{
		cookieName:        "ubclaunchpad-inertia",
		cookieDomain:      domain,
		cookiePath:        path,
		cookieTimeout:     time.Duration(timeout) * time.Minute,
		usersBucket:       []byte("users"),
		sessions:          make(map[string]*session),
		endSessionCleanup: make(chan bool),
	}

	// Set up database
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(manager.usersBucket)
		return err
	})
	manager.db = db
	go manager.cleanSessions()

	return manager, nil
}

// Close ends the session cleanup job and releases the DB handler
func (m *userManager) Close() error {
	m.endSessionCleanup <- true
	return m.db.Close()
}

// Reset deletes all users and drops all active sessions
func (m *userManager) Reset() error {
	m.sessions = make(map[string]*session)
	return m.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(m.usersBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucket(m.usersBucket)
		return err
	})
}

// AddUser inserts a new user
func (m *userManager) AddUser(username, password string, admin bool) error {
	err := validateCredentialValues(username, password)
	if err != nil {
		return err
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	props := userProps{HashedPassword: string(hashedPassword), Admin: admin}
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		bytes, err := json.Marshal(props)
		if err != nil {
			return err
		}
		return users.Put([]byte(username), bytes)
	})
}

// RemoveUser removes user with given username and ends related sessions
func (m *userManager) RemoveUser(username string) error {
	m.endAllUserSessions(username)
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		return users.Delete([]byte(username))
	})
}

// HasUser returns nil if user exists in database
func (m *userManager) HasUser(username string) error {
	found := false
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		user := users.Get([]byte(username))
		if user != nil {
			found = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !found {
		return errUserNotFound
	}
	return nil
}

// IsCorrectCredentials checks if username and password has a match
// in the database
func (m *userManager) IsCorrectCredentials(username, password string) (bool, error) {
	correct := false
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes == nil {
			return errUserNotFound
		}

		props := &userProps{}
		err := json.Unmarshal(propsBytes, props)
		if err != nil {
			return errors.New("Corrupt user properties: " + err.Error())
		}
		correct = correctPassword(props.HashedPassword, password)
		return nil
	})
	return correct, err
}

// IsAdmin checks if given user is has administrator priviledges
func (m *userManager) IsAdmin(username string) (bool, error) {
	admin := false
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			err := json.Unmarshal(propsBytes, props)
			if err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}
			admin = props.Admin
		}
		return nil
	})
	return admin, err
}

// SessionBegin starts a new session with user by setting a cookie
// and adding session to memory
func (m *userManager) SessionBegin(username string, w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(m.cookieTimeout)
	id := generateSessionID()
	s := &session{
		Username: username,
		Expires:  expiration,
	}
	m.sessions[id] = s
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    url.QueryEscape(id),
		Domain:   m.cookieDomain,
		Path:     m.cookiePath,
		HttpOnly: true,
		Expires:  expiration,
	})
}

// SessionEnd ends a session and sets cookie to expire
func (m *userManager) SessionEnd(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	id, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return
	}
	delete(m.sessions, id)
	expiration := time.Now()
	newCookie := http.Cookie{
		Name:     m.cookieName,
		Domain:   m.cookieDomain,
		Path:     m.cookiePath,
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	}
	http.SetCookie(w, &newCookie)
}

// GetSession verifies if given request is from a valid session and returns it
func (m *userManager) GetSession(w http.ResponseWriter, r *http.Request) (*session, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, errCookieNotFound
	}
	id, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}
	s, found := m.sessions[id]
	if !found {
		return nil, errSessionNotFound
	}
	if !m.isValidSession(s) {
		m.SessionEnd(w, r)
		return nil, errSessionNotFound
	}
	return s, nil
}

// endAllUserSessions removes all active sessions with given user
func (m *userManager) endAllUserSessions(username string) {
	for id, s := range m.sessions {
		if s.Username == username {
			delete(m.sessions, id)
		}
	}
}

// isValidSession checks if session is expired
func (m *userManager) isValidSession(s *session) bool {
	return s.Expires.After(time.Now())
}

// cleanSessions is a goroutine that continously cleans sessions
func (m *userManager) cleanSessions() {
	for {
		select {
		case <-m.endSessionCleanup:
			return
		default:
			for id, s := range m.sessions {
				if !m.isValidSession(s) {
					delete(m.sessions, id)
				}
			}
			time.AfterFunc(
				time.Duration(m.cookieTimeout)*time.Minute,
				func() { m.cleanSessions() },
			)
		}
	}
}
