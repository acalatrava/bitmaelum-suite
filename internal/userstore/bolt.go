// Copyright (c) 2020 BitMaelum Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package userstore

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type boltRepo struct {
	client *bolt.DB
}

const (
	keyNotFound       string = "key not found"
	userstoreNotFound string = "userstore not found"
)

// BoltDBFile is the filename to store the boltdb database
const BoltDBFile = "userstore.db"

// NewBoltRepository initializes a new repository
func NewBoltRepository(dbpath string) Repository {
	dbFile := filepath.Join(dbpath, BoltDBFile)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		logrus.Error("Unable to open filepath ", dbFile, err)
		return nil
	}

	return boltRepo{
		client: db,
	}
}

/*
func getEntries(tx *bolt.Tx, c *bolt.Cursor) []StoreEntry {
	var entries []StoreEntry

	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			newBucket := c.Bucket().Bucket([]byte(key))
			if newBucket == nil {
				newBucket = tx.Bucket()
			}
			entries = append(entries, getEntries(tx, newBucket.Cursor()))
		} else {
			entry := StoreEntry{}
			json.Unmarshal(v, &entry)
			entries = append(entries, entry)
		}
	}
}
*/

// Dump the whole store
func (b boltRepo) Dump(addr hash.Hash) (*[]StoreEntry, error) {
	return dump(b, false, addr)
}

// Dump the whole store
func (b boltRepo) DumpIndex(addr hash.Hash) (*[]StoreEntry, error) {
	return dump(b, true, addr)
}

func dump(b boltRepo, onlyIndex bool, addr hash.Hash) (*[]StoreEntry, error) {
	var entries []StoreEntry

	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		c := userBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			entry := StoreEntry{}
			if onlyIndex {
				if !entry.IsCollection {
					continue
				}
			}
			json.Unmarshal(v, &entry)
			entries = append(entries, entry)
		}

		return nil
	})

	return &entries, err
}

// Fetch the given key
func (b boltRepo) Fetch(addr hash.Hash, key string) (*StoreEntry, error) {
	entry := &StoreEntry{}

	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}
		/*
			mainBucket := userBucket.Bucket([]byte(key))
			if mainBucket == nil {
				logrus.Trace("key not found: ", key)
				return errors.New(keyNotFound)
			}

			v := mainBucket.Get([]byte(key))
		*/
		v := userBucket.Get([]byte(key))
		if v == nil {
			return errors.New(keyNotFound)
		}

		err := json.Unmarshal(v, &entry)
		return err
	})

	return entry, err
}

// Store the given key in the repository
func (b boltRepo) Store(addr hash.Hash, entry StoreEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = b.client.Update(func(tx *bolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(addr.Byte())
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", addr.String(), err)
			return err
		}

		return userBucket.Put([]byte(entry.ID), data)
	})

	return err
}

// Remove the given key from the repository
func (b boltRepo) Remove(addr hash.Hash, key string) error {
	return b.client.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("unable to delete entry, bucket on BOLT: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		return userBucket.Delete([]byte(key))
	})
}
