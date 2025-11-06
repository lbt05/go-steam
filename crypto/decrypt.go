package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

// SymmetricDecrypt decrypts data with AES-256 where the first 16 bytes (encrypted with ECB)
// hold the IV. It then uses that IV to decrypt the rest of the data in CBC mode.
// If checkHmac is true, it verifies that the IV (which is structured as a partial HMAC
// concatenated with 3 random bytes) matches the HMAC computed over the plaintext.
func SymmetricDecrypt(input, key []byte, checkHmac bool) ([]byte, error) {
	switch {
	case len(input) < aes.BlockSize:
		return nil, fmt.Errorf("input too short")

	case len(key) != 32:
		return nil, fmt.Errorf("key must be 32 bytes")
	}

	// Create the AES cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt the IV.
	iv := make([]byte, aes.BlockSize)
	block.Decrypt(iv, input[:aes.BlockSize])

	// The ciphertext should be a multiple of the block size.
	var ciphertext = input[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Create the CBC decrypter.
	cbcDecrypter := cipher.NewCBCDecrypter(block, iv)
	plaintextPadded := make([]byte, len(ciphertext))
	cbcDecrypter.CryptBlocks(plaintextPadded, ciphertext)

	// Remove PKCS7 padding.
	plaintext, err := Pkcs7Unpad(plaintextPadded, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad plaintext: %w", err)
	}

	// If the HMAC check is enabled, verify the HMAC.
	if checkHmac {
		if len(iv) < 3 {
			return nil, fmt.Errorf("IV too short for HMAC check")
		}

		// The remote partial HMAC is the first 13 bytes of the IV.
		remotePartialHmac := iv[:len(iv)-3]

		// The random part is the last 3 bytes of the IV.
		randomPart := iv[len(iv)-3:]

		// Compute the HMAC.
		mac := hmac.New(sha1.New, key[:16])
		mac.Write(randomPart)
		mac.Write(plaintext)
		computedDigest := mac.Sum(nil)

		if !hmac.Equal(remotePartialHmac, computedDigest[:len(remotePartialHmac)]) {
			return nil, fmt.Errorf("received invalid HMAC from remote host")
		}
	}

	return plaintext, nil
}

// SymmetricDecryptECB decrypts data that was encrypted using AES-256 in ECB mode with PKCS7 padding.
func SymmetricDecryptECB(input, key []byte) ([]byte, error) {
	switch {
	case len(input) == 0:
		return nil, fmt.Errorf("input cannot be empty")

	case len(key) != 32:
		return nil, fmt.Errorf("key must be 32 bytes")
	}

	// Create the AES cipher.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// The input should be a multiple of the block size.
	if len(input)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("input is not a multiple of block size")
	}

	// Decrypt the data.
	var decrypted = make([]byte, len(input))
	for bs, be := 0, block.BlockSize(); bs < len(input); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Decrypt(decrypted[bs:be], input[bs:be])
	}

	// Return the decrypted data.
	return Pkcs7Unpad(decrypted, block.BlockSize())
}
