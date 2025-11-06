package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	_ "embed"
)

//go:embed public-key.pem
var steamPublicKeyBytes []byte

// steamPublicKey is the RSA public key used to verify the authenticity of the Steam server.
var steamPublicKey *rsa.PublicKey

// init initializes the steamPublicKey variable by parsing the steamPublicKeyBytes.
func init() {
	var block, _ = pem.Decode(steamPublicKeyBytes)
	if block == nil {
		panic("failed to decode steam public key pem")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("failed to parse steam public key: %w", err))
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		panic(fmt.Errorf("expected RSA public key, got %T", key))
	}

	steamPublicKey = rsaKey
}
