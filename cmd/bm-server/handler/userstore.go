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

package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bitmaelum/bitmaelum-suite/internal/container"

	"github.com/bitmaelum/bitmaelum-suite/internal/userstore"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/gorilla/mux"
)

const (
	keyIsMissing string = "key is missing"
)

type inputStoreEntry struct {
	Data         string `json:"data"`
	Parent       string `json:"parent"`
	IsCollection bool   `json:"iscollection"`
}

// RetrieveStore will retrieve a key or collection
func RetrieveStore(w http.ResponseWriter, req *http.Request) {
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	onlyIndex := true
	dump := req.URL.Query().Get("dump")
	if strings.ToLower(dump) == "true" {
		onlyIndex = false
	}

	logrus.Trace("RetrieveStore called for addr ", h, " and key ", k)

	entries, err := dumpStore(onlyIndex, h, k)
	if err != nil {
		msg := fmt.Sprintf("error while retrieving store: %s", err)
		ErrorOut(w, http.StatusInternalServerError, msg)
		return
	}

	// Output entries
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(entries)

	/*
		if k == "" {
			logrus.Trace("Trying to dump keys")
			entries, err := dumpStore(onlyIndex, h, k)
			if err != nil {
				msg := fmt.Sprintf("error while retrieving store: %s", err)
				ErrorOut(w, http.StatusInternalServerError, msg)
				return
			}

			// Output entries
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(entries)

			return
		}

		repo := container.GetUserStoreRepo()
		entry, err := repo.Fetch(h, k)
		if err != nil {
			msg := fmt.Sprintf("error while fetching key: %s", err)
			ErrorOut(w, http.StatusInternalServerError, msg)
			return
		}

		if entry.IsCollection {
			logrus.Trace("Trying to dump keys for key ", k)
			entries, err := dumpStore(onlyIndex, h, k)
			if err != nil {
				msg := fmt.Sprintf("error while retrieving store: %s", err)
				ErrorOut(w, http.StatusInternalServerError, msg)
				return
			}

			// Output entries
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(entries)

			return
		}

		logrus.Trace("Entry retrieved ", entry.Data)

		// Output entry
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(entry)
	*/
}

// UpdateStore will update a key or collection
func UpdateStore(w http.ResponseWriter, req *http.Request) {
	// Get variables from request
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	// Decode post body
	var input inputStoreEntry
	err = DecodeBody(w, req.Body, &input)
	if err != nil {
		ErrorOut(w, http.StatusBadRequest, "incorrect body")
		return
	}

	data, err := base64.StdEncoding.DecodeString(input.Data)
	if err != nil {
		ErrorOut(w, http.StatusBadRequest, "incorrect data")
		return
	}

	repo := container.GetUserStoreRepo()

	// Check if belongs to a collection
	if input.Parent != "" {
		parentEntry, err := repo.Fetch(h, input.Parent)
		if err != nil {
			ErrorOut(w, http.StatusBadRequest, "parent not found")
			return
		}

		// Check if parent is actually a collection
		if !parentEntry.IsCollection {
			ErrorOut(w, http.StatusBadRequest, "parent is not a collection")
			return
		}
	}

	// Generate new entry and store
	entry := userstore.NewEntry(k, data, input.Parent, input.IsCollection)
	err = repo.Store(h, entry)
	if err != nil {
		ErrorOut(w, http.StatusInternalServerError, "unable to store the data")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveStore will remove a key or collection
func RemoveStore(w http.ResponseWriter, req *http.Request) {
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	repo := container.GetUserStoreRepo()
	if k == "" {
		// Delete the whole store
		entries, err := repo.Dump(h, k)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, accountNotFound)
			return
		}

		for _, entry := range *entries {
			if entry.IsCollection {
				deleteCollection(repo, &h, &entry)
			} else {
				entry.Data = []byte("*deleted*")
				repo.Store(h, entry)
				//repo.Remove(h, entry.ID)
			}
		}

	} else {
		key, _ := hash.NewFromHash(k)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, keyIsMissing)
			return
		}

		entry, err := repo.Fetch(h, key.String())
		if err != nil {
			msg := fmt.Sprintf("error while fetching key: %s", err)
			ErrorOut(w, http.StatusInternalServerError, msg)
			return
		}

		if entry.IsCollection {
			deleteCollection(repo, &h, entry)
		} else {
			entry.Data = []byte("*deleted*")
			repo.Store(h, *entry)
			//repo.Remove(h, key.String())
		}

	}

	// All ok
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func getVariables(req *http.Request) (hash.Hash, string, error) {
	k := mux.Vars(req)["key"]
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		return "", "", errors.New("accountNotFound")
	}

	return *h, k, nil
}

func deleteCollection(repo userstore.Repository, addr *hash.Hash, entry *userstore.StoreEntry) {
	for _, k := range entry.Entries {
		newEntry, err := repo.Fetch(*addr, k)
		if err != nil {
			continue
		}

		if newEntry.IsCollection {
			deleteCollection(repo, addr, newEntry)
		} else {
			newEntry.Data = []byte("*deleted*")
			repo.Store(*addr, *newEntry)
			//repo.Remove(*addr, k)
		}
	}

	entry.Data = []byte("*deleted*")
	entry.IsCollection = false
	repo.Store(*addr, *entry)
	//repo.Remove(*addr, entry.ID)
}

type mNode struct {
	Value    []byte            `json:"data,omitempty"`
	Children map[string]*mNode `json:"children,omitempty"`
}

func dumpStore(onlyIndex bool, addr hash.Hash, key string) (interface{}, error) {
	repo := container.GetUserStoreRepo()

	var (
		entries *[]userstore.StoreEntry
		err     error
	)

	if onlyIndex {
		entries, err = repo.DumpIndex(addr, key)
	} else {
		entries, err = repo.Dump(addr, key)
	}

	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for _, entry := range *entries {
		logrus.Trace("entry -> ", entry.Parent, " ", entry.ID)

		if m[entry.Parent] == nil {
			m[entry.Parent] = make(map[string]interface{})
		}

		if m[entry.ID] == nil {
			if entry.IsCollection {
				m[entry.ID] = make(map[string]interface{})
			} else {
				m[entry.ID] = entry.Data
			}
		} else {
			// This check is because key may have been changed from collection to deleted
			switch m[entry.ID].(type) {
			case []byte:
				if !entry.IsCollection {
					m[entry.ID] = entry.Data
				}
			}
		}

		m[entry.Parent].(map[string]interface{})[entry.ID] = m[entry.ID]
	}

	/*
		for _, entry := range *entries {
			if entry.IsCollection {
				node := make(map[string]interface{})
				m[entry.ID] = node
			} else {
				m[entry.ID] = entry.Data
			}
		}

		for _, entry := range *entries {
			logrus.Trace("checking entry  ", entry.ID, " parent ", entry.Parent)
			if entry.Parent == "" {
				logrus.Trace("entry -> root ", entry.ID)
				m["root"].(map[string]interface{})[entry.ID] = m[entry.ID]
			} else {
				logrus.Trace("entry -> ", entry.Parent, " ", entry.ID)
				m[entry.Parent].(map[string]interface{})[entry.ID] = m[entry.ID]
			}
		}
	*/

	if m[key] == nil {
		// Nothing to be returned
		return m, nil
	}

	return m[key], nil
}
