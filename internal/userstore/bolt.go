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
	parentNotFound    string = "parent key not found"
	userstoreNotFound string = "userstore not found"
	rootDummyKey      string = "root"
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
func (b boltRepo) Dump(addr hash.Hash, key string, since int64) (*[]StoreEntry, error) {
	return dump(b, false, addr, key, since)
}

// Dump the whole store
func (b boltRepo) DumpIndex(addr hash.Hash, key string, since int64) (*[]StoreEntry, error) {
	return dump(b, true, addr, key, since)
}

func dump(b boltRepo, onlyIndex bool, addr hash.Hash, key string, since int64) (*[]StoreEntry, error) {
	var entries []StoreEntry
	var err error

	if key == "" { // Get all entries, faster iteration
		err = b.client.View(func(tx *bolt.Tx) error {
			userBucket := tx.Bucket(addr.Byte())
			if userBucket == nil {
				logrus.Trace("userstore not found for address: ", addr.String())
				return errors.New(userstoreNotFound)
			}

			c := userBucket.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				entry := StoreEntry{}
				json.Unmarshal(v, &entry)
				logrus.Trace("id: ", entry.ID)

				// Do not return "root" dummy key
				if entry.ID == rootDummyKey {
					continue
				}

				// If requested since ignore older entries
				if since > entry.TimeStamp {
					continue
				}

				if onlyIndex {
					if entry.Data != nil && !entry.IsCollection {
						continue
					}
				}
				entries = append(entries, entry)
			}

			return nil
		})
	} else {
		entries, err = getEntriesAndDescendants(b, onlyIndex, addr, key, since)
	}

	logrus.Trace("entries: ", entries)
	return &entries, err
}

func getEntriesAndDescendants(b boltRepo, onlyIndex bool, addr hash.Hash, key string, since int64) ([]StoreEntry, error) {
	var entries []StoreEntry
	var entry StoreEntry

	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		v := userBucket.Get([]byte(key))
		if v == nil {
			return errors.New(keyNotFound)
		}

		return json.Unmarshal(v, &entry)
	})

	if err != nil {
		return nil, err
	}

	// If requested since ignore older entries
	if since < entry.TimeStamp {
		if onlyIndex {
			if entry.IsCollection || entry.Data == nil {
				entries = append(entries, entry)
			}
		} else {
			entries = append(entries, entry)
		}
	}

	if len(entry.Entries) > 0 {
		for _, e := range entry.Entries {
			moreEntries, err := getEntriesAndDescendants(b, onlyIndex, addr, e, since)
			if err != nil {
				return nil, err
			}
			entries = append(entries, moreEntries...)
		}
	}

	return entries, nil
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
		v := userBucket.Get([]byte(key))
		if v == nil {
			return errors.New(keyNotFound)
		}

		err := json.Unmarshal(v, &entry)
		return err
	})

	return entry, err
}

func createRootIfNeeded(b boltRepo, addr hash.Hash) error {
	rootEntry := &StoreEntry{}

	logrus.Trace("createRootIfNeeded")

	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		v := userBucket.Get([]byte(rootDummyKey))
		if v == nil {
			return errors.New(parentNotFound)
		}

		err := json.Unmarshal(v, &rootEntry)
		return err
	})

	if err != nil {
		logrus.Trace("root entry not found, creating it")
		rootEntry := NewEntry(rootDummyKey, nil, "", true)
		data, err := json.Marshal(rootEntry)
		if err != nil {
			return err
		}

		err = b.client.Update(func(tx *bolt.Tx) error {
			userBucket, err := tx.CreateBucketIfNotExists(addr.Byte())
			if err != nil {
				logrus.Trace("unable to create bucket on BOLT: ", addr.String(), err)
				return err
			}

			return userBucket.Put([]byte(rootEntry.ID), data)
		})

		if err != nil {
			logrus.Trace("error creating root ", err)
		}
		return err
	}

	if err != nil {
		logrus.Trace("error retrieving root ", err)
	}

	return err
}

func updateParentsTimestamp(b boltRepo, addr hash.Hash, entry StoreEntry) error {
	// Get parent to update timestamp
	parentEntry := &StoreEntry{}
	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		parent := entry.Parent
		if parent == "" {
			parent = rootDummyKey
		}

		v := userBucket.Get([]byte(parent))
		if v == nil {
			logrus.Trace("parent not found ", parent)
			return errors.New(parentNotFound)
		}

		err := json.Unmarshal(v, &parentEntry)
		return err
	})

	if err != nil {
		return err
	}

	parentEntry.TimeStamp = entry.TimeStamp

	parentData, err := json.Marshal(parentEntry)
	if err != nil {
		return err
	}

	err = b.client.Update(func(tx *bolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(addr.Byte())
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", addr.String(), err)
			return err
		}

		return userBucket.Put([]byte(parentEntry.ID), parentData)
	})

	if err != nil {
		logrus.Trace("unable to update timestamp for ", parentEntry.ID, " -> ", err)
		return err
	}

	if parentEntry.Parent != "" {
		return updateParentsTimestamp(b, addr, *parentEntry)
	}

	return nil
}

func updateParentsChildren(b boltRepo, addr hash.Hash, entry StoreEntry) error {
	// Get parent to update children
	parentEntry := &StoreEntry{}
	err := b.client.View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket(addr.Byte())
		if userBucket == nil {
			logrus.Trace("userstore not found for address: ", addr.String())
			return errors.New(userstoreNotFound)
		}

		parent := entry.Parent
		if parent == "" {
			parent = rootDummyKey
		}

		v := userBucket.Get([]byte(parent))
		if v == nil {
			logrus.Trace("parent not found ", parent)
			return errors.New(parentNotFound)
		}

		err := json.Unmarshal(v, &parentEntry)
		return err
	})

	if err != nil {
		return err
	}

	childInParent := false
	for _, child := range parentEntry.Entries {
		if child == entry.ID {
			childInParent = true
			break
		}
	}

	if !childInParent {
		logrus.Trace("adding ", entry.ID, " to parent childs")
		parentEntry.Entries = append(parentEntry.Entries, entry.ID)
	}

	parentData, err := json.Marshal(parentEntry)
	if err != nil {
		return err
	}

	err = b.client.Update(func(tx *bolt.Tx) error {
		userBucket, err := tx.CreateBucketIfNotExists(addr.Byte())
		if err != nil {
			logrus.Trace("unable to create bucket on BOLT: ", addr.String(), err)
			return err
		}

		return userBucket.Put([]byte(parentEntry.ID), parentData)
	})

	if err != nil {
		logrus.Trace("unable to store updated parent")
	}
	return err
}

// Store the given key in the repository
func (b boltRepo) Store(addr hash.Hash, entry StoreEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	err = createRootIfNeeded(b, addr)
	if err != nil {
		logrus.Trace("unable to create root")
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

	if err != nil {
		logrus.Trace("unable to store entry ", entry.ID)
		return err
	}

	err = updateParentsTimestamp(b, addr, entry)

	if err != nil {
		return err
	}

	err = updateParentsChildren(b, addr, entry)

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
