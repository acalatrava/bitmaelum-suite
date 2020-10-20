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
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/bitmaelum/bitmaelum-suite/internal/container"

	"github.com/bitmaelum/bitmaelum-suite/internal/userstore"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/gorilla/mux"
)

const (
	keyIsMissing string = "key is missing"
)

type parentToEntrySliceMap map[string][]userstore.StoreEntry
type parentToIndexItemMap map[string]map[int]userstore.StoreEntry

type inputStoreEntry struct {
	Data         string `json:"data"`
	Parent       string `json:"parent"`
	IsCollection bool   `json:"iscollection"`
}

func dumpStore(addr hash.Hash, key string) (parentToIndexItemMap, error) {
	repo := container.GetUserStoreRepo()
	entries, err := repo.Dump(addr)
	if err != nil {
		return nil, err
	}

	// collect items according to parent
	parToItemSlice := parentToEntrySliceMap{}
	for _, v := range *entries {
		if v.Parent == key || key == "" {
			parToItemSlice[v.Parent] = append(parToItemSlice[v.Parent], v)
		}
	}

	//turn those slices into int -> Item maps for decoding
	parToIndexItemMap := parentToIndexItemMap{}
	for k, v := range parToItemSlice {
		if parToIndexItemMap[k] == nil {
			parToIndexItemMap[k] = map[int]userstore.StoreEntry{}
		}
		for index, item := range v {
			parToIndexItemMap[k][index] = item
		}
	}

	return parToIndexItemMap, nil
}

// RetrieveStore will retrieve a key or collection
func RetrieveStore(w http.ResponseWriter, req *http.Request) {
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	k := mux.Vars(req)["key"]

	logrus.Trace("RetrieveStore called for addr ", h, " and key ", k)

	if k == "" {
		logrus.Trace("Trying to dump keys")
		entries, err := dumpStore(*h, k)
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
	entry, err := repo.Fetch(*h, k)
	if err != nil {
		msg := fmt.Sprintf("error while fetching key: %s", err)
		ErrorOut(w, http.StatusInternalServerError, msg)
		return
	}

	if entry.IsCollection {
		logrus.Trace("Trying to dump keys for key ", k)
		entries, err := dumpStore(*h, k)
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
}

// UpdateStore will update a key or collection
func UpdateStore(w http.ResponseWriter, req *http.Request) {
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	k := mux.Vars(req)["key"]

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

	entry := userstore.NewEntry(k, data, input.Parent, input.IsCollection)

	repo := container.GetUserStoreRepo()
	err = repo.Store(*h, entry)
	if err != nil {
		ErrorOut(w, http.StatusInternalServerError, "unable to store the data")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteCollection(repo userstore.Repository, addr *hash.Hash, entry *userstore.StoreEntry) {
	for _, k := range entry.Entries {
		newEntry, err := repo.Fetch(*addr, k)
		if err != nil {
			continue
		}

		if entry.IsCollection {
			deleteCollection(repo, addr, newEntry)
		} else {
			repo.Remove(*addr, k)
		}
	}
}

// RemoveStore will remove a key or collection
func RemoveStore(w http.ResponseWriter, req *http.Request) {
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	k, err := hash.NewFromHash(mux.Vars(req)["key"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, keyIsMissing)
		return
	}

	repo := container.GetUserStoreRepo()
	entry, err := repo.Fetch(*h, k.String())
	if err != nil {
		msg := fmt.Sprintf("error while fetching key: %s", err)
		ErrorOut(w, http.StatusInternalServerError, msg)
		return
	}

	if entry.IsCollection {
		deleteCollection(repo, h, entry)
	} else {
		repo.Remove(*h, k.String())
	}

	// All ok
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
