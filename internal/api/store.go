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

package api

import (
	"encoding/base64"
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
)

// PutDataInStore upload data to the store
func (api *API) PutDataInStore(addr hash.Hash, k string, v string, p string) error {
	type inputStoreData struct {
		Data         string `json:"data"`
		Parent       string `json:"parent"`
		IsCollection bool   `json:"iscollection"`
	}

	data := base64.StdEncoding.EncodeToString([]byte(v))
	isCollection := false

	if v == "" {
		data = ""
		isCollection = true
	}

	input := &inputStoreData{
		Data:         data,
		Parent:       p,
		IsCollection: isCollection,
	}

	resp, statusCode, err := api.PutJSON(getURL(addr, k), input)
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode > 299 {
		return getErrorFromResponse(resp)
	}

	return nil
}

// DeleteKeyFromStore delete a key from the store
func (api *API) DeleteKeyFromStore(addr hash.Hash, k string) error {

	resp, statusCode, err := api.Delete(getURL(addr, k))
	if err != nil {
		return err
	}

	if statusCode < 200 || statusCode > 299 {
		return getErrorFromResponse(resp)
	}

	return nil
}

// GetKeyFromStore gets a key data from the store
func (api *API) GetKeyFromStore(addr hash.Hash, k string, dump bool, since string) (interface{}, error) {
	//var entries interface{}

	//resp, statusCode, err := api.GetJSON(fmt.Sprintf("/store/%s/%s", addr.String(), hash.New(k).String()), entries)
	url := getURL(addr, k)

	if dump || since != "" {
		url = url + "?"
	}

	if dump {
		url = url + "dump=true&"
	}

	//@TODO: do this better
	if since != "" {
		url = url + "since=" + since + "000000000" //unix to unixnano
	}

	resp, statusCode, err := api.Get(url)
	if err != nil {
		return nil, err
	}

	jsonParsed, err := gabs.ParseJSON(resp)
	if err != nil {
		return nil, err
	}

	if statusCode < 200 || statusCode > 299 {
		return nil, getErrorFromResponse(resp)
	}

	return jsonParsed, nil
}

func getURL(addr hash.Hash, k string) string {
	var url string

	if k == "" {
		url = fmt.Sprintf("/store/%s", addr.String())
	} else {
		url = fmt.Sprintf("/store/%s/%s", addr.String(), hash.New(k).String())
	}

	return url
}
