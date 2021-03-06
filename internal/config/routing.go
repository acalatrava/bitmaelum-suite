package config

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
	"github.com/bitmaelum/bitmaelum-suite/pkg/hash"
	"github.com/spf13/afero"
)

// RoutingConfig holds routing configuration for the mail server
type RoutingConfig struct {
	RoutingID  string           `json:"routing_id"`
	PrivateKey bmcrypto.PrivKey `json:"private_key"`
	PublicKey  bmcrypto.PubKey  `json:"public_key"`
}

// Routing keeps the routing ID and keys
var Routing RoutingConfig

// ReadRouting will read the routing file and merge it into the server configuration
func ReadRouting(p string) error {
	f, err := fs.Open(p)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	Routing = RoutingConfig{}
	err = json.Unmarshal(data, &Routing)
	if err != nil {
		return err
	}

	return nil
}

// SaveRouting will save the routing into a file. It will overwrite if exists
func SaveRouting(p string, routing *RoutingConfig) error {
	data, err := json.MarshalIndent(routing, "", "  ")
	if err != nil {
		return err
	}

	err = fs.MkdirAll(filepath.Dir(p), 0755)
	if err != nil {
		return err
	}

	return afero.WriteFile(fs, p, data, 0600)
}

// GenerateRouting generates a new routing structure
func GenerateRouting() (string, *RoutingConfig, error) {
	seed, privKey, pubKey, err := bmcrypto.GenerateKeypairWithSeed()
	if err != nil {
		return "", nil, err
	}

	id := hex.EncodeToString(pubKey.K.(ed25519.PublicKey))
	return seed, &RoutingConfig{
		RoutingID:  hash.New(id).String(),
		PrivateKey: *privKey,
		PublicKey:  *pubKey,
	}, nil
}
