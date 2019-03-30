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
		if err != nil {
			return fmt.Errorf("failed to created env variable bucket: %s", err.Error())
		}

		_, err = tx.CreateBucketIfNotExists(deployedProjectsBucket)
		if err != nil {
			return fmt.Errorf("failed to created deployed projects bucket: %s", err.Error())
		}
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

// TODO: Change name, error check, only insert project mdata inside private helper 'update build'
// AddProjectBuildData stores and tracks metadata from successful builds
func (c *DeploymentDataManager) AddProjectBuildData(projectName string, mdata DeploymentMetadata) error {
	// encode metadata so it can be stored as byte array
	encodedMdata, err := json.Marshal(mdata)
	if err != nil {
		return fmt.Errorf("failure encrypting metadata: %s", err.Error())
	}
	err = c.db.Update(func(tx *bolt.Tx) error {
		depProjectsBkt := tx.Bucket(deployedProjectsBucket)
		// if bkt with project name doesnt exist create new bkt, otherwise update existing bucket
		if projectBkt := depProjectsBkt.Bucket([]byte(projectName)); projectBkt == nil {
			projectBkt, err := depProjectsBkt.CreateBucket([]byte(projectName))
			if err != nil {
				return fmt.Errorf("failure creating project bkt: %s", err.Error())
			}

			if err := projectBkt.Put([]byte(time.Now().String()), encodedMdata); err != nil {
				return fmt.Errorf("failure inserting project metadata: %s", err.Error())
			}
		}
		return nil
	})
	return c.UpdateProjectBuildData(projectName, mdata)
}

// UpdateProjectBuildData updates existing project bkt with recent build's metadata
func (c *DeploymentDataManager) UpdateProjectBuildData(projectName string,
	mdata DeploymentMetadata) error {
	// encode metadata so it can be stored as byte array
	encodedMdata, err := json.Marshal(mdata)
	if err != nil {
		return fmt.Errorf("failure encrypting metadata: %s", err.Error())
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		depProjectBkt := tx.Bucket(deployedProjectsBucket)
		projectBkt := depProjectBkt.Bucket([]byte(projectName))

		if err := projectBkt.Put([]byte(time.Now().String()), encodedMdata); err != nil {
			return fmt.Errorf("failure updating db with project metadata: %s", err.Error())
		}
		return nil
	})

}

// GetNumOfDeployedProjects returns number of projects currently deployed
func (c *DeploymentDataManager) GetNumOfDeployedProjects(projectName string) (int, error) {
	var numBkts int
	err := c.db.View(func(tx *bolt.Tx) error {
		depProjectBkt := tx.Bucket(deployedProjectsBucket)
		bktStats := depProjectBkt.Stats()
		numBkts = bktStats.BucketN
		return nil
	})
	return numBkts, err
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
