package auth

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

// userProps are properties associated with user, used
// for database entries
type userProps struct {
	HashedPassword string `json:"hashedPassword"`
	Admin          bool   `json:"admin"`
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
}

func newUserManager(dbPath string, timeout int64) (*userManager, error) {
	manager := &userManager{
		cookieName:     "ubclaunchpad/inertia",
		cookieTimeout:  timeout,
		usersBucket:    []byte("users"),
		sessionsBucket: []byte("sessions"),
	}

	// Set up database
	err := manager.initializeDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *userManager) initializeDatabase(path string) error {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(m.usersBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(m.sessionsBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *userManager) Close() error {
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

func (m *userManager) SessionBegin(username string) {
}

func (m *userManager) SessionEnd(username string) {
}

func (m *userManager) SessionDestroy(username string) {
}
