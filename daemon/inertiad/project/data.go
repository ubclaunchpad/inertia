package project

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

var (
	// database buckets
	envVariableBucket = []byte("envVariables")
)

// DataManager stores persistent deployment configuration
type DataManager struct {
	// db is a boltdb database, which is an embedded
	// key/value database where each "bucket" is a collection
	db *bolt.DB

	// @TODO: should these keys be here?
	// Keys for encrypting data
	encryptPublicKey  *[32]byte
	encryptPrivateKey *[32]byte
	// Keys for decrypting data
	decryptPublicKey  *[32]byte
	decryptPrivateKey *[32]byte
}

func newDataManager(dbPath string) (*DataManager, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(envVariableBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	encryptPublicKey, encryptPrivateKey, decryptPublicKey, decryptPrivateKey, err := crypto.GenerateKeys()

	return &DataManager{
		db,
		encryptPublicKey, encryptPrivateKey,
		decryptPublicKey, decryptPrivateKey,
	}, nil
}

// AddEnvVariable adds a new environment variable that will be applied
// to all project containers
func (c *DataManager) AddEnvVariable(name, value string,
	encrypt bool) error {
	if len(name) == 0 || len(value) == 0 {
		return errors.New("invalid env configuration")
	}

	valueBytes := []byte(value)
	if encrypt {
		encrypted, err := crypto.Seal(valueBytes,
			c.encryptPrivateKey, c.decryptPublicKey)
		if err != nil {
			return err
		}
		valueBytes = encrypted
	}

	return c.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket(envVariableBucket)
		bytes, err := json.Marshal(envVariable{
			Value:     valueBytes,
			Encrypted: encrypt,
		})
		if err != nil {
			return err
		}
		return users.Put([]byte(name), bytes)
	})
}

// RemoveEnvVariable removes a previously set env variable
func (c *DataManager) RemoveEnvVariable(name string) error {
	return nil
}

// GetEnvVariables retrieves all stored environment variables
func (c *DataManager) GetEnvVariables(decrypt bool) (map[string]string, error) {
	env := map[string]string{}

	err := c.db.View(func(tx *bolt.Tx) error {
		variables := tx.Bucket(envVariableBucket)
		return variables.ForEach(func(name, variableBytes []byte) error {
			variable := &envVariable{}
			err := json.Unmarshal(variableBytes, variable)
			if err != nil {
				return err
			}

			if !variable.Encrypted {
				env[string(name)] = string(variable.Value)
			} else if !decrypt {
				env[string(name)] = "[ENCRYPTED]"
			} else {
				decrypted, err := crypto.UndoSeal(variable.Value,
					c.encryptPublicKey, c.decryptPrivateKey)
				if err != nil {
					return err
				}
				env[string(name)] = string(decrypted)
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return env, nil
}

func (c *DataManager) destroy() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(envVariableBucket)
	})
}
