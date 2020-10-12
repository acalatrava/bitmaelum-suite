package client

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/bitmaelum/bitmaelum-suite/internal/container"
	"github.com/bitmaelum/bitmaelum-suite/internal/message"
	"github.com/bitmaelum/bitmaelum-suite/pkg/bmcrypto"
)

// SignHeader will add a server signature to a message header. This can be used to proof the origin of the message
func SignHeader(header *message.Header, privKey bmcrypto.PrivKey) error {
	// Already signed? Then skip
	if len(header.ClientSignature) > 0 {
		fmt.Println("already signed")
		return nil
	}

	data, err := json.Marshal(header)
	if err != nil {
		return err
	}

	h := sha256.Sum256(data)
	sig, err := bmcrypto.Sign(privKey, h[:])
	if err != nil {
		return err
	}

	header.ClientSignature = base64.StdEncoding.EncodeToString(sig)
	return nil
}

// VerifyHeader will verify a server signature from a message header. This can be used to proof the origin of the message
func VerifyHeader(header message.Header) bool {
	// Fetch public key from routing
	rs := container.GetResolveService()

	addr, err := rs.ResolveAddress(header.From.Addr)
	if err != nil {
		return false
	}

	// No header at all
	if len(header.ClientSignature) == 0 {
		return false
	}

	// Store signature
	targetSignature, err := base64.StdEncoding.DecodeString(header.ClientSignature)
	if err != nil {
		return false
	}
	header.ServerSignature = ""
	header.ClientSignature = ""

	// Generate hash
	data, err := json.Marshal(&header)
	if err != nil {
		return false
	}
	h := sha256.Sum256(data)

	// Verify signature
	ok, err := bmcrypto.Verify(addr.PublicKey, h[:], []byte(targetSignature))
	if err != nil {
		return false
	}

	return ok
}
