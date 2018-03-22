package auth

import (
	"errors"

	"github.com/boltdb/bolt"
)

// userManager administers sessions and user accounts
type userManager struct {
	db            *bolt.DB
	userBucket    []byte
	adminBucket   []byte
	sessionBucket []byte
}

func newUserManager(dbPath string) (*userManager, error) {
	manager := &userManager{
		userBucket:    []byte("users"),
		adminBucket:   []byte("admins"),
		sessionBucket: []byte("sessions"),
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
		_, err = tx.CreateBucketIfNotExists(m.userBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(m.adminBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(m.sessionBucket)
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

func (m *userManager) AddUser(username, password string) error {
	err := validateCredentialValues(username, password)
	if err != nil {
		return err
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.userBucket)
		return users.Put([]byte(username), []byte(hashedPassword))
	})
}

func (m *userManager) RemoveUser(username string) error {
	return nil
}

func (m *userManager) AssignAdmin(username string) error {
	return nil
}

// Check Functions

func (m *userManager) HasUser(username string) (bool, error) {
	found := false
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.userBucket)
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
		users := tx.Bucket(m.userBucket)
		hashedPassword := users.Get([]byte(username))
		if hashedPassword == nil {
			return errors.New("User not found")
		}
		correct = correctPassword(hashedPassword, password)
		return nil
	})
	if err != nil {
		return false, err
	}
	return correct, nil
}

func (m *userManager) IsAdmin(username string) (bool, error) {
	return false, nil
}

// Session Management

func (m *userManager) LogIn(username string) {
	return
}

func (m *userManager) LogOut(username string) {
	return
}
