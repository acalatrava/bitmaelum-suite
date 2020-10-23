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
	"time"

	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// StoreEntry represents an entry data to the store
type StoreEntry struct {
	ID             string   `json:"id"`             // The key of the entry
	Parent         string   `json:"parent"`         // What is the parent collection (hash-key) of this collection
	IsCollection   bool     `json:"collection"`     // If this entry is a collection
	Data           []byte   `json:"data"`           // The blob data of the collection
	TimeStamp      int64    `json:"timestamp"`      // The timestamp of last modification time
	Entries        []string `json:"entries"`        // The entries that belongs to this entry (if this is a collection)
	Subcollections []string `json:"subcollections"` // The subcollections underneath this collection
}

// Repository is a repository to fetch and store StoreEntry entries
type Repository interface {
	DumpIndex(addr hash.Hash, key string) (*[]StoreEntry, error)
	Dump(addr hash.Hash, key string) (*[]StoreEntry, error)
	Fetch(addr hash.Hash, key string) (*StoreEntry, error)
	Store(addr hash.Hash, entry StoreEntry) error
	Remove(addr hash.Hash, key string) error
}

// NewEntry will create a new entry
func NewEntry(key string, data []byte, parent string, iscol bool) StoreEntry {
	return StoreEntry{
		ID:             key,
		Parent:         parent,
		IsCollection:   iscol,
		Data:           data,
		TimeStamp:      time.Now().UnixNano(),
		Entries:        nil,
		Subcollections: nil,
	}
}
