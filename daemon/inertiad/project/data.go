package project

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"

	"github.com/boltdb/bolt"
	"golang.org/x/crypto/nacl/box"
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

	encryptPublicKey, encryptPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	decryptPublicKey, decryptPrivateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

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

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}

	// This encrypts the variable, storing the nonce in the first 24 bytes.
	if encrypt {
		variable := []byte(valueBytes)
		valueBytes = box.Seal(
			nonce[:], variable, &nonce,
			c.decryptPublicKey, c.encryptPrivateKey,
		)
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
func (c *DataManager) RemoveEnvVariable(name, value string) error {
	return nil
}

func (c *DataManager) getEnvVariables() (map[string]string, error) {
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
			} else {
				// Decrypt the message using decrypt private key and the
				// encrypt public key. When you decrypt, you must use the same
				// nonce you used to encrypt the message - this nonce is stored
				// in the first 24 bytes.
				var decryptNonce [24]byte
				copy(decryptNonce[:], variable.Value[:24])
				decrypted, ok := box.Open(
					nil, variable.Value[24:], &decryptNonce,
					c.encryptPublicKey, c.decryptPrivateKey,
				)
				if !ok {
					return errors.New("decryption error")
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
		return nil
	})
}
