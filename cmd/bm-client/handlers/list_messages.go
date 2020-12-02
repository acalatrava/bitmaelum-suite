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
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/internal"
	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/cmd/bm-client/internal/mailbox"
	"github.com/bitmaelum/bitmaelum-suite/internal/api"
	"github.com/bitmaelum/bitmaelum-suite/internal/message"
	"github.com/bitmaelum/bitmaelum-suite/internal/vault"
	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/c2h5oh/datasize"
	"github.com/olekukonko/tablewriter"
)

// ListMessages will display message information from accounts and boxes
func ListMessages(accounts []vault.AccountInfo, since time.Time) {
	table := tablewriter.NewWriter(os.Stdout)
	// table.SetAutoMergeCells(true)

	headers := []string{"Account", "Box", "ID", "Subject", "From", "Date", "# Blocks", "# Attachments"}
	table.SetHeader(headers)

	msgCount := 0

	fmt.Print("* Fetching remote messages...")
	for _, info := range accounts {
		// Fetch routing info
		resolver := container.Instance.GetResolveService()
		routingInfo, err := resolver.ResolveRouting(info.RoutingID)
		if err != nil {
			continue
		}

		client, err := api.NewAuthenticated(*info.Address, &info.PrivKey, routingInfo.Routing, internal.JwtErrorFunc)
		if err != nil {
			continue
		}

		msgCount += displayBoxList(client, &info, table, since)
	}
	fmt.Println("")

	if msgCount == 0 {
		fmt.Println("* No messages in the given timespan")
	} else {
		table.Render()
	}
}

var firstRender bool

func displayBoxList(client *api.API, account *vault.AccountInfo, table *tablewriter.Table, since time.Time) int {
	mbl, err := client.GetMailboxList(account.Address.Hash())
	if err != nil {
		return 0
	}

	msgCount := 0
	firstRender = true
	for _, mb := range mbl.Boxes {
		msgCount += displayBox(client, account, fmt.Sprintf("%d", mb.ID), table, since)
	}

	return msgCount
}

func displayBox(client *api.API, account *vault.AccountInfo, box string, table *tablewriter.Table, since time.Time) int {
	mb, err := client.GetMailboxMessages(account.Address.Hash(), box, since)
	if err != nil {
		return 0
	}

	// No messages in this box found
	if len(mb.Messages) == 0 {
		return 0
	}

	// Display account info on first render
	if firstRender {
		firstRender = false

		values := []string{
			account.Address.String(),
			box,
			"", "", "", "", "", "",
		}
		table.Append(values)
	} else {
		values := []string{
			"",
			box,
			"", "", "", "", "", "",
		}
		table.Append(values)
	}

	// Sort messages first
	msort := mailbox.NewMessageSort(&account.PrivKey, mb.Messages, mailbox.SortDate, true)
	sort.Sort(&msort)

	for _, msg := range mb.Messages {
		key, err := bmcrypto.Decrypt(account.PrivKey, msg.Header.Catalog.TransactionID, msg.Header.Catalog.EncryptedKey)
		if err != nil {
			continue
		}
		catalog := &message.Catalog{}
		err = bmcrypto.CatalogDecrypt(key, msg.Catalog, catalog)
		if err != nil {
			continue
		}

		var blocks []string
		for _, b := range catalog.Blocks {
			blocks = append(blocks, b.Type)
		}

		var attachments []string
		for _, a := range catalog.Attachments {
			fs := datasize.ByteSize(a.Size)
			attachments = append(attachments, a.FileName+" ("+fs.HR()+")")
		}

		values := []string{
			"",
			"",
			msg.ID,
			catalog.Subject,
			fmt.Sprintf("%s <%s>", catalog.From.Name, catalog.From.Address),
			catalog.CreatedAt.Format(time.RFC822),
			strings.Join(blocks, ","),
			strings.Join(attachments, "\n"),
		}

		table.Append(values)
	}

	return len(mb.Messages)
}