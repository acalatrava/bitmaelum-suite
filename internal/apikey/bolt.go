package apikey

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
)

type boltRepo struct {
	client *bolt.DB
}

//BucketName is the bucket name to store the invitations on the bolt db
const BucketName = "apikeys"

//BoltDBFile is the filename to store the boltdb database
const BoltDBFile = "apikeys.db"

// NewBoltRepository initializes a new repository
func NewBoltRepository(dbpath *string) Repository {
	dbFile := filepath.Join(*dbpath, BoltDBFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		logrus.Error("Unable to open filepath ", dbFile, err)
		return nil
	}

	return boltRepo{
		client: db,
	}
}

// Fetch a key from the repository, or err
func (b boltRepo) Fetch(ID string) (*KeyType, error) {
	key := &KeyType{}

	err := b.client.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			logrus.Trace("apikey not found in BOLT: ", ID, nil)
			return errors.New("apikey not found")
		}

		data := bucket.Get([]byte(ID))
		if data == nil {
			logrus.Trace("apikey not found in BOLT: ", data, nil)
			return errors.New("apikey not found")
		}

		err := json.Unmarshal([]byte(data), &key)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return key, nil
}

// Store the given key in the repository
func (b boltRepo) Store(apiKey KeyType) error {
	data, err := json.Marshal(apiKey)
	if err != nil {
		return err
	}

	err = b.client.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", BucketName, err)
			return err
		}

		return bucket.Put([]byte(apiKey.ID), data)
	})

	return err
}

// Remove the given key from the repository
func (b boltRepo) Remove(ID string) {

	_ = b.client.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		if bucket == nil {
			logrus.Trace("unable to delete apikey, apikey not found in BOLT: ", ID, nil)
			return nil
		}

		return bucket.Delete([]byte(ID))
	})

}
