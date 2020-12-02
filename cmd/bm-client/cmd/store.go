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

package cmd

import (
	"os"

	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/handlers"
	"github.com/bitmaelum/bitmaelum-suite/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var storeCmd = &cobra.Command{
	Use:     "store",
	Aliases: []string{"store", "st"},
	Short:   "Manage the server storage",
	Long:    `It will allow to retrieve, store and delete encrypted arbitrary data to the server`,
	Run: func(cmd *cobra.Command, args []string) {
		if !*storePut && !*storeGet && !*storeDel {
			logrus.Fatalf("please specify at least one action (upload, get, remove)")
			os.Exit(1)
		}

		if *storePut && *storeKey == "" {
			logrus.Fatalf("please specify key")
			os.Exit(1)
		}

		if *storeGet && *storeData != "" {
			logrus.Fatalf("data and get cannot be used together")
			os.Exit(1)
		}

		if *storeGet && *storeParent != "" {
			logrus.Fatalf("parent and get cannot be used together")
			os.Exit(1)
		}

		v := vault.OpenVault()

		accountToUse := vault.GetAccountOrDefault(v, *storeAccount)
		if accountToUse == nil {
			logrus.Fatal("No account found in vault")
			os.Exit(1)
		}

		// Fetch routing info
		resolver := container.Instance.GetResolveService()
		routingInfo, err := resolver.ResolveRouting(accountToUse.RoutingID)
		if err != nil {
			logrus.Fatal("Cannot find routing ID for this account")
			os.Exit(1)
		}

		if *storePut {
			handlers.StorePut(accountToUse, routingInfo, storeKey, storeData, storeParent)
		}

		if *storeDel {
			handlers.StoreDel(accountToUse, routingInfo, storeKey)
		}

		if *storeGet {
			handlers.StoreGet(accountToUse, routingInfo, storeKey, storeDump, storeSince)
		}
	},
}

var (
	storeAccount *string
	storeData    *string
	storeKey     *string
	storeParent  *string
	storeSince   *string
	storePut     *bool
	storeGet     *bool
	storeDel     *bool
	storeDump    *bool
)

func init() {
	rootCmd.AddCommand(storeCmd)

	storeAccount = storeCmd.Flags().StringP("account", "a", "", "Account to use")
	storeData = storeCmd.Flags().StringP("data", "d", "", "Arbitrary data to be stored")
	storeKey = storeCmd.Flags().StringP("key", "k", "", "Key ID of the data")
	storeParent = storeCmd.Flags().String("parent", "", "Parent key ID where this data belongs to")
	storePut = storeCmd.Flags().BoolP("upload", "u", false, "Store the data to the server")
	storeGet = storeCmd.Flags().BoolP("get", "g", false, "Retrieve the key from the server or, if requested a collection, the underlaying structure")
	storeDel = storeCmd.Flags().BoolP("remove", "r", false, "Remove the key from the server")
	storeDump = storeCmd.Flags().BoolP("dump", "", false, "Dump the whole structure and return all the values")
	storeSince = storeCmd.Flags().String("since", "", "Unix timestamp to retrieve only updated items")

	//_ = storeCmd.MarkFlagRequired("key")
}