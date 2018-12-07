package project

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

var (
	// database buckets
	envVariableBucket = []byte("envVariables")
)

// DeploymentDataManager stores persistent deployment configuration
type DeploymentDataManager struct {
	// db is a boltdb database, which is an embedded
	// key/value database where each "bucket" is a collection
	db *bolt.DB

	// @TODO: should these keys be here?
	// Keys for encrypting data
	symmetricKey []byte
}

func newDataManager(dbPath string) (*DeploymentDataManager, error) {
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

	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return nil, fmt.Errorf("Failed to generate key: %s", key)
	}
	return &DeploymentDataManager{
		db,
		key,
	}, nil
}

// AddEnvVariable adds a new environment variable that will be applied
// to all project containers
func (c *DeploymentDataManager) AddEnvVariable(name, value string,
	encrypt bool) error {
	if len(name) == 0 || len(value) == 0 {
		return errors.New("invalid env configuration")
	}

	valueBytes := []byte(value)
	if encrypt {
		encrypted, err := crypto.Encrypt(c.symmetricKey, valueBytes)
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
func (c *DeploymentDataManager) RemoveEnvVariable(name string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		vars := tx.Bucket(envVariableBucket)
		return vars.Delete([]byte(name))
	})
}

// GetEnvVariables retrieves all stored environment variables
func (c *DeploymentDataManager) GetEnvVariables(decrypt bool) ([]string, error) {
	envs := []string{}

	err := c.db.View(func(tx *bolt.Tx) error {
		variables := tx.Bucket(envVariableBucket)
		return variables.ForEach(func(name, variableBytes []byte) error {
			variable := &envVariable{}
			err := json.Unmarshal(variableBytes, variable)
			if err != nil {
				return err
			}

			nameString := string(name)
			if !variable.Encrypted {
				envs = append(envs, nameString+"="+string(variable.Value))
			} else if !decrypt {
				envs = append(envs, nameString+"=[ENCRYPTED]")
			} else {
				decrypted, err := crypto.Decrypt(variable.Value, c.symmetricKey)
				if err != nil {
					// If decrypt fails, key is no longer valid - remove var
					c.RemoveEnvVariable(nameString)
				}
				envs = append(envs, nameString+"="+string(decrypted))
			}

			return nil
		})
	})
	return envs, err
}

func (c *DeploymentDataManager) destroy() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(envVariableBucket); err != nil {
			return err
		}
		_, err := tx.CreateBucket(envVariableBucket)
		return err
	})
}
