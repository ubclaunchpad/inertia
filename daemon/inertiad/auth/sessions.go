package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

type sessionManager struct {
	// sessionTimeout is the amount of time created Tokens are given to expire
	sessionTimeout time.Duration

	// internal is sessionManager's session store - it is protected by an RWMutex
	internal map[string]*crypto.TokenClaims
	sync.RWMutex

	// keyLookup implements jwt.Keyfunc and retrieves the key used to sign and
	// validate JWT tokens
	keyLookup func(*jwt.Token) (interface{}, error)

	// endSessionCleanup ends the goroutine that continually cleans up expired
	// essions from memory
	endSessionCleanup chan bool
}

func newSessionManager(domain string, timeout int,
	keyLookup func(*jwt.Token) (interface{}, error)) *sessionManager {
	manager := &sessionManager{
		sessionTimeout: time.Duration(timeout) * time.Minute,
		internal:       make(map[string]*crypto.TokenClaims),
		keyLookup:      keyLookup,

		endSessionCleanup: make(chan bool),
	}

	// Set up session cleanup goroutine
	ticker := time.NewTicker(manager.sessionTimeout)
	go func() {
		for {
			select {
			case <-manager.endSessionCleanup:
				ticker.Stop()
				return
			case <-ticker.C:
				manager.Lock()
				for id, c := range manager.internal {
					if c.Valid() != nil {
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
	s.internal = make(map[string]*crypto.TokenClaims)
	s.Unlock()
}

// SessionBegin starts a new session with user by generating a token and adding
// session to memory
func (s *sessionManager) BeginSession(username string, admin bool) (*crypto.TokenClaims, string, error) {
	expiration := time.Now().Add(s.sessionTimeout)
	id, err := common.GenerateRandomString()
	if err != nil {
		return nil, "", fmt.Errorf("Faield to begin sesison for %s: %s", username, err.Error())
	}

	claims := &crypto.TokenClaims{
		SessionID: id, User: username, Admin: admin, Expiry: expiration,
	}

	// Sign a token for user
	keyBytes, err := s.keyLookup(nil)
	if err != nil {
		return nil, "", err
	}
	token, err := claims.GenerateToken(keyBytes.([]byte))
	if err != nil {
		return nil, "", err
	}

	// Add session to map
	s.Lock()
	s.internal[id] = claims
	s.Unlock()
	return claims, token, nil
}

// SessionEnd ends a session by invalidating the token
func (s *sessionManager) EndSession(r *http.Request) error {
	claims, err := s.GetSession(r)
	if err != nil {
		return err
	}

	// Delete session from map
	s.deleteSession(claims.SessionID)
	return nil
}

// GetSession verifies if given request is from a valid session and returns it
func (s *sessionManager) GetSession(r *http.Request) (*crypto.TokenClaims, error) {
	// Check token in header - if no tokens, check cookie
	bearerString := r.Header.Get("Authorization")
	splitToken := strings.Split(bearerString, "Bearer ")
	if len(splitToken) != 2 {
		return nil, errors.New(errMalformedHeaderMsg)
	}

	// Validate token and get claims
	claims, err := crypto.ValidateToken(splitToken[1], s.keyLookup)
	if err != nil {
		return nil, err
	}

	// Master tokens aren't session-tracked. TODO: reassess security of this
	if claims.IsMaster() {
		return claims, nil
	}

	s.RLock()
	_, found := s.internal[claims.SessionID]
	if !found || claims.Valid() != nil {
		s.RUnlock()
		s.deleteSession(claims.SessionID)
		return nil, errSessionNotFound
	}
	s.RUnlock()
	return claims, nil
}

// endAllUserSessions removes all active sessions with given user
func (s *sessionManager) EndAllUserSessions(username string) {
	for id, claim := range s.internal {
		if claim.User == username {
			s.deleteSession(id)
		}
	}
}

// EndAllSessions removes all active sessions
func (s *sessionManager) EndAllSessions() {
	s.Lock()
	s.internal = make(map[string]*crypto.TokenClaims)
	s.Unlock()
}

func (s *sessionManager) deleteSession(sessionID string) {
	s.Lock()
	delete(s.internal, sessionID)
	s.Unlock()
}
