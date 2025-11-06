package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"fmt"
)

// GenerateSessionKey generates a 32-byte symmetric session key and encrypts it with Steam's public key.
// It returns the plain session key and the encrypted session key.
func GenerateSessionKey() ([]byte, []byte, error) {
	// Generate a 32-byte symmetric session key.
	var plainSessionKey = make([]byte, 32)
	if _, err := rand.Read(plainSessionKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	// Encrypt the session key with Steam's public key.
	encryptedSessionKey, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, steamPublicKey, plainSessionKey, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt session key: %w", err)
	}

	return plainSessionKey, encryptedSessionKey, nil
}
