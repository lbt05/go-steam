package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
)

// SymmetricEncrypt encrypts the input using a symmetric key and an optional IV.
// If an IV is not provided, a new one is generated.
// The IV is used to encrypt the input in CBC mode.
// The input is padded with PKCS7 padding.
// The function returns the encrypted data.
func SymmetricEncrypt(input, key, iv []byte) ([]byte, error) {
	switch {
	case len(input) == 0:
		return nil, fmt.Errorf("input cannot be empty")

	case len(key) != 32:
		return nil, fmt.Errorf("key must be 32 bytes")

	case iv != nil && len(iv) != aes.BlockSize:
		return nil, fmt.Errorf("IV must be %d bytes", aes.BlockSize)
	}

	// If an IV is not provided, generate a new one.
	if iv == nil {
		iv = make([]byte, aes.BlockSize)
		if _, err := rand.Read(iv); err != nil {
			return nil, fmt.Errorf("failed to generate IV: %w", err)
		}
	}

	// Create the AES cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Encrypt the IV.
	encryptedIV := make([]byte, aes.BlockSize)
	block.Encrypt(encryptedIV, iv)

	// Apply PKCS7 padding.
	paddedInput := Pkcs7Pad(input, aes.BlockSize)

	// Create the CBC encrypter.
	cbcEncrypter := cipher.NewCBCEncrypter(block, iv)
	encryptedData := make([]byte, len(paddedInput))
	cbcEncrypter.CryptBlocks(encryptedData, paddedInput)

	// Return the encrypted data.
	return append(encryptedIV, encryptedData...), nil
}

// SymmetricEncryptWithHmacIv encrypts the input using a symmetric key and a custom IV.
// The IV is computed as follows:
//
//	IV = HMAC-SHA1( random(3) || input )[0:(16-3)] || random(3)
//
// That is, generate 3 random bytes, compute HMAC-SHA1 over random bytes and input
// using the first 16 bytes of key, then take the first 13 bytes of the digest and
// append the random bytes to form a 16-byte IV.
func SymmetricEncryptWithHmacIv(input, key []byte) ([]byte, error) {
	switch {
	case len(input) == 0:
		return nil, fmt.Errorf("input cannot be empty")

	case len(key) < 16:
		return nil, fmt.Errorf("key must be at least 16 bytes")
	}

	// Generate 3 random bytes.
	var randomBytes = make([]byte, 3)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Compute HMAC-SHA1 using the first 16 bytes of the key.
	mac := hmac.New(sha1.New, key[:16])
	mac.Write(randomBytes)
	mac.Write(input)

	// Compute the digest.
	digest := mac.Sum(nil)
	iv := append(digest[:16-len(randomBytes)], randomBytes...)

	return SymmetricEncrypt(input, key, iv)
}
