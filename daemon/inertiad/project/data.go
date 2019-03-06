package project

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	bolt "go.etcd.io/bbolt"
)

var (
	// database buckets
	envVariableBucket      = []byte("envVariables")
	deployedProjectsBucket = []byte("deployedProjects")
)

// DeploymentDataManager stores persistent deployment configuration
type DeploymentDataManager struct {
	// db is a boltdb database, which is an embedded
	// key/value database where each "bucket" is a collection
	db *bolt.DB

	// Keys for encrypting data
	symmetricKey []byte
}

// NewDataManager instantiates a database associated with a deployment
func NewDataManager(dbPath string, keyPath string) (*DeploymentDataManager, error) {
	// retrieve AES key, generate if not present
	var key []byte
	var err error
	if key, err = ioutil.ReadFile(keyPath); err != nil || len(key) != crypto.SymmetricKeyLength {
		key = make([]byte, crypto.SymmetricKeyLength)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate key: %s", key)
		}
		os.Remove(keyPath)
		if err := ioutil.WriteFile(keyPath, key, 0600); err != nil {
			return nil, fmt.Errorf("failed to write key to '%s': %s", keyPath, err.Error())
		}
	}

	// Set up database
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at '%s': %s", dbPath, err.Error())
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(envVariableBucket)
		_, err = tx.CreateBucketIfNotExists(deployedProjectsBucket)
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to instantiate database: %s", err.Error())
	}

	return &DeploymentDataManager{
		db,
		key,
	}, nil
}

// AddEnvVariable adds a new environment variable that will be applied
// to all project containers
func (c *DeploymentDataManager) AddEnvVariable(name, value string, encrypt bool) error {
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
		vars := tx.Bucket(envVariableBucket)
		bytes, err := json.Marshal(envVariable{
			Value:     valueBytes,
			Encrypted: encrypt,
		})
		if err != nil {
			return err
		}
		return vars.Put([]byte(name), bytes)
	})
}

// RemoveEnvVariables removes previously set env variables
func (c *DeploymentDataManager) RemoveEnvVariables(names ...string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		var vars = tx.Bucket(envVariableBucket)
		for _, n := range names {
			if err := vars.Delete([]byte(n)); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetEnvVariables retrieves all stored environment variables
func (c *DeploymentDataManager) GetEnvVariables(decrypt bool) ([]string, error) {
	var envs = []string{}
	var faulty = []string{}
	var err = c.db.View(func(tx *bolt.Tx) error {
		var variables = tx.Bucket(envVariableBucket)
		return variables.ForEach(func(name, variableBytes []byte) error {
			var variable = &envVariable{}
			if err := json.Unmarshal(variableBytes, variable); err != nil {
				return err
			}

			var nameString = string(name)
			if !variable.Encrypted {
				envs = append(envs, nameString+"="+string(variable.Value))
			} else if !decrypt {
				envs = append(envs, nameString+"=[ENCRYPTED]")
			} else {
				decrypted, err := crypto.Decrypt(c.symmetricKey, variable.Value)
				if err != nil {
					// If decrypt fails, key is no longer valid - remove var
					faulty = append(faulty, nameString)
				}
				envs = append(envs, nameString+"="+string(decrypted))
			}
			return nil
		})
	})

	c.RemoveEnvVariables(faulty...)

	return envs, err
}

// AddProjectBuildData stores and tracks metadata from successful builds
func (c *DeploymentDataManager) AddProjectBuildData(projectName string, mdata DeploymentMetadata) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		deployedProjectsBucket := tx.Bucket(deployedProjectsBucket)
		// if project bkt doesnt exist create new bkt, otherwise update existing bucket
		if projectBkt := deployedProjectsBucket.Bucket([]byte(projectName)); projectBkt == nil {
			projectBkt, err := tx.CreateBucket([]byte(projectName))
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(mdata)
			if err != nil {
				return err
			}
			projectBkt.Put([]byte(time.Now().String()), encoded)
		}
		return nil
	})
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
