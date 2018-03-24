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
	errUserNotFound    = errors.New("User not found")
)

// userProps are properties associated with user, used
// for database entries
type userProps struct {
	HashedPassword string `json:"hashedPassword"`
	Admin          bool   `json:"admin"`
}

// sessionProps are properties associated with session,
// used for database entries
type sessionProps struct {
	Username string    `json:"username"`
	Expires  time.Time `json:"created"`
}

// userManager administers sessions and user accounts
type userManager struct {
	cookieName    string
	cookieTimeout int64

	// bolt.DB is an embedded key/value database,
	// where each "bucket" is a collection
	db             *bolt.DB
	usersBucket    []byte
	sessionsBucket []byte

	endSessionCleanup chan bool
}

func newUserManager(dbPath string, timeout int64) (*userManager, error) {
	manager := &userManager{
		cookieName:     "ubclaunchpad/inertia",
		cookieTimeout:  timeout,
		usersBucket:    []byte("users"),
		sessionsBucket: []byte("sessions"),
	}

	// Set up database
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(manager.usersBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(manager.sessionsBucket)
		if err != nil {
			return err
		}

		return nil
	})
	manager.db = db

	manager.endSessionCleanup = make(chan bool)
	go manager.cleanSessions()

	return manager, nil
}

func (m *userManager) Close() error {
	m.endSessionCleanup <- true
	return m.db.Close()
}

// User Administration Functions

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

func (m *userManager) RemoveUser(username string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		return users.Delete([]byte(username))
	})
}

// User Checks

func (m *userManager) HasUser(username string) (bool, error) {
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
		return true, err
	}
	return found, nil
}

func (m *userManager) IsCorrectCredentials(username, password string) (bool, error) {
	correct := false
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes == nil {
			return errors.New("User not found")
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

// Session Management

func (m *userManager) SessionBegin(username string, w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		id := generateSessionID()
		expiration := time.Now().Add(time.Duration(m.cookieTimeout) * time.Minute)
		err := m.addSession(id, username, expiration)
		if err != nil {
			return err
		}
		cookie := http.Cookie{
			Name:     m.cookieName,
			Value:    url.QueryEscape(id),
			Path:     "/web",
			Expires:  expiration,
			HttpOnly: true,
		}
		http.SetCookie(w, &cookie)
	}
	return nil
}

func (m *userManager) SessionEnd(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	m.removeSession(cookie.Value)
	expiration := time.Now()
	newCookie := http.Cookie{
		Name:     m.cookieName,
		Path:     "/web",
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	}
	http.SetCookie(w, &newCookie)
}

func (m *userManager) SessionCheck(w http.ResponseWriter, r *http.Request) (bool, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return false, err
	}
	s, err := m.getSession(cookie.Value)
	if err != nil {
		return false, err
	}
	return m.isValidSession(s), nil
}

// Session Helpers

func (m *userManager) removeSession(id string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(m.sessionsBucket)
		return sessions.Delete([]byte(id))
	})
}

func (m *userManager) addSession(id, username string, expires time.Time) error {
	props := sessionProps{Username: username, Expires: expires}
	return m.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(m.sessionsBucket)
		bytes, err := json.Marshal(props)
		if err != nil {
			return err
		}
		return sessions.Put([]byte(id), bytes)
	})
}

func (m *userManager) getSession(id string) (*sessionProps, error) {
	props := &sessionProps{}
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.sessionsBucket)
		propsBytes := users.Get([]byte(id))
		if propsBytes != nil {
			err := json.Unmarshal(propsBytes, props)
			if err != nil {
				return errors.New("Corrupt session properties: " + err.Error())
			}
		} else {
			return errSessionNotFound
		}
		return nil
	})
	return props, err
}

func (m *userManager) isValidSession(s *sessionProps) bool {
	return s.Expires.Before(time.Now())
}

// cleanSessions is a goroutine that continously cleans sessions
func (m *userManager) cleanSessions() {
	for {
		select {
		case <-m.endSessionCleanup:
			return
		default:
			m.db.Update(func(tx *bolt.Tx) error {
				sessions := tx.Bucket(m.sessionsBucket)
				c := sessions.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					s := &sessionProps{}
					err := json.Unmarshal(v, s)
					if err != nil {
						println(err)
						continue
					}
					if !m.isValidSession(s) {
						err = sessions.Delete(k)
						if err != nil {
							println(err)
						}
					}
				}
				return nil
			})
			time.AfterFunc(
				time.Duration(m.cookieTimeout)*time.Minute,
				func() { m.cleanSessions() },
			)
		}
	}
}
