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

type node struct {
<<<<<<< HEAD
	ID       string `json:"id"`
	Value    []byte `json:"data,omitempty"`
	Children *node  `json:"children,omitempty"`
	ParentID string `json:"parent"`
=======
	ID       string               `json:"id"`
	Value    userstore.StoreEntry `json:"value"`
	Children []node               `json:"children,omitempty"`
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
}

type parentToEntrySliceMap map[string][]userstore.StoreEntry
type parentToIndexItemMap map[string]map[int]userstore.StoreEntry

type inputStoreEntry struct {
	Data         string `json:"data"`
	Parent       string `json:"parent"`
	IsCollection bool   `json:"iscollection"`
}

<<<<<<< HEAD
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
=======
func addToTree(root []node, entries []userstore.StoreEntry) []node {
	if len(entries) > 0 {
		var i int
		for i = 0; i < len(root); i++ {
			if root[i].ID == entries[0].ID { //already in tree
				break
			}
		}
		if i == len(root) {
			root = append(root, node{ID: entries[0].ID})
		}
		root[i].Children = addToTree(root[i].Children, entries[1:])
	}
	return root
}

func dumpStore(addr hash.Hash, key string) ([]node, error) {
	var tree []node

	repo := container.GetUserStoreRepo()
	entries, err := repo.Dump(addr)
	if err != nil {
		return tree, err
	}

	tree = addToTree(tree, *entries)

	/*
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
	*/

	return tree, nil
}

// RetrieveStore will retrieve a key or collection
func RetrieveStore(w http.ResponseWriter, req *http.Request) {
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	k := mux.Vars(req)["key"]
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f

	logrus.Trace("RetrieveStore called for addr ", h, " and key ", k)

	if k == "" {
		logrus.Trace("Trying to dump keys")
<<<<<<< HEAD
		entries, err := dumpStore(onlyIndex, h, k)
=======
		entries, err := dumpStore(*h, k)
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
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
<<<<<<< HEAD
	entry, err := repo.Fetch(h, k)
=======
	entry, err := repo.Fetch(*h, k)
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
	if err != nil {
		msg := fmt.Sprintf("error while fetching key: %s", err)
		ErrorOut(w, http.StatusInternalServerError, msg)
		return
	}

	if entry.IsCollection {
		logrus.Trace("Trying to dump keys for key ", k)
<<<<<<< HEAD
		entries, err := dumpStore(onlyIndex, h, k)
=======
		entries, err := dumpStore(*h, k)
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
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
<<<<<<< HEAD
	// Get variables from request
	h, k, err := getVariables(req)
	if err != nil {
		ErrorOut(w, http.StatusNotFound, err.Error())
		return
	}

	// Decode post body
=======
	h, err := hash.NewFromHash(mux.Vars(req)["addr"])
	if err != nil {
		ErrorOut(w, http.StatusNotFound, accountNotFound)
		return
	}

	k := mux.Vars(req)["key"]

>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
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

<<<<<<< HEAD
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
=======
	entry := userstore.NewEntry(k, data, input.Parent, input.IsCollection)

	repo := container.GetUserStoreRepo()
	err = repo.Store(*h, entry)
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
	if err != nil {
		ErrorOut(w, http.StatusInternalServerError, "unable to store the data")
		return
	}

	w.WriteHeader(http.StatusOK)
}

<<<<<<< HEAD
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
		entries, err := repo.Dump(h)
		if err != nil {
			ErrorOut(w, http.StatusNotFound, accountNotFound)
			return
		}

		for _, entry := range *entries {
			if entry.IsCollection {
				deleteCollection(repo, &h, &entry)
			} else {
				repo.Remove(h, entry.ID)
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
			repo.Remove(h, key.String())
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

=======
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
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
<<<<<<< HEAD

	repo.Remove(*addr, entry.ID)
}

func dumpStore(onlyIndex bool, addr hash.Hash, key string) ([]node, error) {
	tree := []node{}
	tree2 := []node{}

	repo := container.GetUserStoreRepo()

	var (
		entries *[]userstore.StoreEntry
		err     error
	)

	if onlyIndex {
		entries, err = repo.DumpIndex(addr)
	} else {
		entries, err = repo.Dump(addr)
	}

	if err != nil {
		return tree, err
	}

	a, _ := json.Marshal(entries)
	err = json.Unmarshal(a, &tree)
	if err != nil {
		return tree, err
	}

	m := make(map[string]*node)
	for i, entry := range *entries {
		//fmt.Printf("Setting m[%d] = <node with ID=%d>\n", n.ID, n.ID)
		m[entry.ID] = &tree[i]
	}

	for i, n := range tree {
		//fmt.Printf("Setting <node with ID=%d>.Child to <node with ID=%d>\n", n.ID, m[n.ParentID].ID)
		if m[n.ParentID] != nil {
			m[n.ParentID].Children = &tree[i]
		}
	}

	for _, n := range tree {
		if (n.ParentID == key && key != "") || key == "" {
			tree2 = append(tree2, n)
		}
	}

	/*
		tree = addToTree(tree, *entries)
	*/
	/*
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
	*/

	return tree2, nil
}

func dumpIndex(addr hash.Hash, key string) ([]node, error) {
	tree := []node{}
	tree2 := []node{}

	repo := container.GetUserStoreRepo()
	entries, err := repo.Dump(addr)

	if err != nil {
		return tree, err
	}

	a, _ := json.Marshal(entries)
	err = json.Unmarshal(a, &tree)
	if err != nil {
		return tree, err
	}

	m := make(map[string]*node)
	for i, entry := range *entries {
		//fmt.Printf("Setting m[%d] = <node with ID=%d>\n", n.ID, n.ID)
		m[entry.ID] = &tree[i]
	}

	for i, n := range tree {
		//fmt.Printf("Setting <node with ID=%d>.Child to <node with ID=%d>\n", n.ID, m[n.ParentID].ID)
		if m[n.ParentID] != nil {
			m[n.ParentID].Children = &tree[i]
		}
	}

	for _, n := range tree {
		if (n.ParentID == key && key != "") || key == "" {
			tree2 = append(tree2, n)
		}
	}

	/*
		tree = addToTree(tree, *entries)
	*/
	/*
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
	*/

	return tree2, nil
=======
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
>>>>>>> 1dabfdaedb07d8a914764c8aaa233b3123d7540f
}
