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

package handlers

import (
	"fmt"

	"github.com/bitmaelum/bitmaelum-suite/internal"
	"github.com/bitmaelum/bitmaelum-suite/internal/api"
	"github.com/bitmaelum/bitmaelum-suite/internal/config"
	"github.com/bitmaelum/bitmaelum-suite/internal/resolver"
	"github.com/bitmaelum/bitmaelum-suite/pkg/address"
	"github.com/sirupsen/logrus"
)

func getClient(info *internal.AccountInfo, routingInfo *resolver.RoutingInfo) *api.API {
	client, err := api.NewAuthenticated(info, api.ClientOpts{
		Host:          routingInfo.Routing,
		AllowInsecure: config.Client.Server.AllowInsecure,
		Debug:         config.Client.Server.DebugHTTP,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	return client
}

// StorePut put data on the store
func StorePut(info *internal.AccountInfo, routingInfo *resolver.RoutingInfo, k *string, v *string, p *string) {
	client := getClient(info, routingInfo)

	addr, _ := address.NewAddress(info.Address)
	addrHash := addr.Hash()

	err := client.PutDataInStore(addrHash, *k, *v, *p)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Done.")
}

// StoreGet gets data from the store
func StoreGet(info *internal.AccountInfo, routingInfo *resolver.RoutingInfo, k *string, dump *bool, since *string) {
	client := getClient(info, routingInfo)

	addr, _ := address.NewAddress(info.Address)
	addrHash := addr.Hash()

	j, err := client.GetKeyFromStore(addrHash, *k, *dump, *since)

	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Print(j)
	fmt.Println("Done.")
}

// StoreDel remove data frome the store
func StoreDel(info *internal.AccountInfo, routingInfo *resolver.RoutingInfo, k *string) {
	client := getClient(info, routingInfo)

	addr, _ := address.NewAddress(info.Address)
	addrHash := addr.Hash()

	err := client.DeleteKeyFromStore(addrHash, *k)

	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println("Done.")
}
