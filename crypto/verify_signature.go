package crypto

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
)

func VerifySignature(algorithm string, data, signature []byte) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("data cannot be empty")
	}
	if len(signature) == 0 {
		return false, fmt.Errorf("signature cannot be empty")
	}

	var hasher crypto.Hash
	switch algorithm {
	case "", "RSA-SHA1":
		hasher = crypto.SHA1

	case "RSA-SHA256":
		hasher = crypto.SHA256

	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	var hash = hasher.New()
	if _, err := hash.Write(data); err != nil {
		return false, fmt.Errorf("failed to write data to hasher: %w", err)
	}

	var digest = hash.Sum(nil)
	switch err := rsa.VerifyPKCS1v15(steamPublicKey, hasher, digest, signature); {
	case errors.Is(err, rsa.ErrVerification):
		return false, nil

	case err != nil:
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	return true, nil
}
