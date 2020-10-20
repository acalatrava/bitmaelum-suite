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

package container

import (
	"sync"

	"github.com/bitmaelum/bitmaelum-suite/internal/config"
	"github.com/bitmaelum/bitmaelum-suite/internal/userstore"
)

var (
	userStoreOnce       sync.Once
	userStoreRepository userstore.Repository
)

// GetUserStoreRepo returns the repository for storing and fetching api keys
func GetUserStoreRepo() userstore.Repository {

	userStoreOnce.Do(func() {
		/*
			//If redis.host is set on the config file it will use redis instead of bolt
			if config.Server.Redis.Host != "" {
				opts := redis.Options{
					Addr: config.Server.Redis.Host,
					DB:   config.Server.Redis.Db,
				}

				userStoreRepository = userstore.NewRedisRepository(&opts)
				return
			}
		*/
		//If redis is not set then it will use BoltDB as default
		userStoreRepository = userstore.NewBoltRepository(config.Server.Bolt.DatabasePath)
	})

	return userStoreRepository
}
