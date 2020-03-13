package auth

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	bolt "go.etcd.io/bbolt"
)

var (
	errUserNotFound       = errors.New("user not found")
	errBackupCodeNotFound = errors.New("backup code not found")
	errMissingCredentials = errors.New("no credentials provided")
)

const (
	masterKey = "master"
)

// userProps are properties associated with user, used
// for database entries
type userProps struct {
	HashedPassword  string
	Admin           bool
	LoginAttempts   int
	TotpSecret      string
	TotpBackupCodes []string
}

// userManager administers sessions and user accounts
type userManager struct {
	// db is a boltdb database, which is an embedded key/value database where
	// each "bucket" is a collection
	db          *bolt.DB
	usersBucket []byte
}

func newUserManager(dbPath string) (*userManager, error) {
	manager := &userManager{
		usersBucket: []byte("users"),
	}

	// Set up database
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		users, err := tx.CreateBucketIfNotExists(manager.usersBucket)
		if err != nil {
			return err
		}
		// Add a master user - the password to this guy/gal will just be the
		// GitHub key. It's not really meant for use.
		bytes, err := json.Marshal(&userProps{Admin: true})
		if err != nil {
			return err
		}
		return users.Put([]byte(masterKey), bytes)
	})
	if err != nil {
		return nil, err
	}
	manager.db = db

	return manager, nil
}

// Close ends the session cleanup job and releases the DB handler
func (m *userManager) Close() error {
	return m.db.Close()
}

// Reset deletes all users and drops all active sessions
func (m *userManager) Reset() error {
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		return users.ForEach(func(username, v []byte) error {
			if string(username) != masterKey {
				if err := users.Delete(username); err != nil {
					tx.Rollback()
					return err
				}
			}
			return nil
		})
	})
}

// AddUser inserts a new user
func (m *userManager) AddUser(username, password string, admin bool) error {
	err := crypto.ValidateCredentialValues(username, password)
	if err != nil {
		return err
	}
	hashedPassword, err := crypto.HashPassword(password)
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
	var u = []byte(username)
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		if users.Get(u) == nil {
			return errUserNotFound
		}
		return users.Delete(u)
	})
}

// UserList returns a list of all registered users
func (m *userManager) UserList() []string {
	userList := make([]string, 0)
	m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		return users.ForEach(func(username, v []byte) error {
			userList = append(userList, string(username))
			return nil
		})
	})
	return userList
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

// IsCorrectCredentials checks if username and password has a match in the database
func (m *userManager) IsCorrectCredentials(username, password string) (*userProps, bool, error) {
	if username == "" || password == "" {
		return nil, false, errMissingCredentials
	}

	var (
		key     = []byte(username)
		props   = &userProps{}
		userErr error
		correct bool
	)

	transactionErr := m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		if propsBytes := users.Get(key); propsBytes == nil {
			return errUserNotFound
		} else if err := json.Unmarshal(propsBytes, props); err != nil {
			return errors.New("Corrupt user properties: " + err.Error())
		}

		// The 'correct' here is returned by the funtion
		correct = crypto.CorrectPassword(props.HashedPassword, password)
		if !correct {
			// Track number of login attempts
			props.LoginAttempts++

			// We went through several iterations of behaviour here, but each one had issues with
			// potential DOS attacks:
			// * exponential backoffs
			// * deleting user after x attempts
			// For now, it seems the best response is to do nothing, and allow unlimited attempts.
			// Eventually, we might want to add some sort of reset mechanism when a limit is reached.
			// We'll maintain the behaviour of tracking login attempts just in case - it might be
			// useful for auditing.
			bytes, err := json.Marshal(props)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
			return users.Put(key, bytes)
		}

		// Reset attempts to 0 if login successful
		props.LoginAttempts = 0
		bytes, err := json.Marshal(props)
		if err != nil {
			return err
		}

		// Put overwrites existing entry to update it
		return users.Put(key, bytes)
	})

	if userErr != nil {
		return props, correct, userErr
	}
	return props, correct, transactionErr
}

// IsValidTotp returns true if the given TOTP is valid for the given user, and
// false otherwise.
func (m *userManager) IsValidTotp(username string, totp string) (bool, error) {
	var totpSecret string
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			if err := json.Unmarshal(propsBytes, props); err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}
			totpSecret = props.TotpSecret
			return nil
		}
		return errors.New("No such user")
	})
	if err != nil {
		return false, err
	}
	return crypto.ValidatePasscode(totp, totpSecret), nil
}

// IsValidBackupCode returns true if the given backup code is valid for the
// given user, and false otherwise.
func (m *userManager) IsValidBackupCode(username, backupCode string) (bool, error) {
	var backupCodes []string
	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			if err := json.Unmarshal(propsBytes, props); err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}
			backupCodes = props.TotpBackupCodes
			return nil
		}
		return errors.New("No such user")
	})
	if err != nil {
		return false, err
	}
	for _, correctBackupCode := range backupCodes {
		if backupCode == correctBackupCode {
			return true, nil
		}
	}

	return false, nil
}

// IsAdmin checks if given user is has administrator priviledges
func (m *userManager) IsAdmin(username string) (bool, error) {
	// Check if user is admin in database
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

// IsTotpEnabled checks if a given user has TOTP enabled
func (m *userManager) IsTotpEnabled(username string) (bool, error) {
	totpEnabled := false

	err := m.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			err := json.Unmarshal(propsBytes, props)
			if err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}
			if props.TotpSecret != "" {
				totpEnabled = true
			}
		}
		return nil
	})
	return totpEnabled, err
}

// EnableTotp enables TOTP for a user
func (m *userManager) EnableTotp(username string) (string, []string, error) {
	props := &userProps{}

	err := m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			err := json.Unmarshal(propsBytes, props)
			if err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}

			totpSecret, totpErr := crypto.GenerateSecretKey(username)
			if totpErr != nil {
				return errors.New("Error generating secret totp key: " + totpErr.Error())
			}
			props.TotpBackupCodes = crypto.GenerateBackupCodes()
			props.TotpSecret = totpSecret.Secret()

			bytes, err := json.Marshal(props)
			if err != nil {
				return err
			} else if err = users.Put([]byte(username), bytes); err != nil {
				return err
			}
		} else {
			return errors.New("Cannot enable totp, user does not exist")
		}
		return nil
	})
	return props.TotpSecret, props.TotpBackupCodes, err
}

// DisableTotp disables TOTP for a user
func (m *userManager) DisableTotp(username string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			err := json.Unmarshal(propsBytes, props)
			if err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}
			props.TotpSecret = ""
			props.TotpBackupCodes = []string{}

			bytes, err := json.Marshal(props)
			if err != nil {
				return err
			}
			err = users.Put([]byte(username), bytes)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Cannot disable totp, user does not exist")
		}
		return nil
	})
}

// RemoveBackupCode removes the given backup code from the user's list of
// backup codes
func (m *userManager) RemoveBackupCode(username, backupCode string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(m.usersBucket)
		propsBytes := users.Get([]byte(username))
		if propsBytes != nil {
			props := &userProps{}
			if err := json.Unmarshal(propsBytes, props); err != nil {
				return errors.New("Corrupt user properties: " + err.Error())
			}

			// find the backup code
			backupCodes := props.TotpBackupCodes
			index := -1
			for i, storedBackupCode := range backupCodes {
				if storedBackupCode == backupCode {
					index = i
					break
				}
			}

			// doesn't exist
			if index == -1 {
				return errBackupCodeNotFound
			}

			// remove it
			props.TotpBackupCodes = append(
				props.TotpBackupCodes[:index],
				props.TotpBackupCodes[index+1:]...)

			// store updated user
			bytes, err := json.Marshal(props)
			if err != nil {
				return err
			}

			return users.Put([]byte(username), bytes)
		}
		return errUserNotFound
	})
}
