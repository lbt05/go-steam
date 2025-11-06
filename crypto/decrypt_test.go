package crypto_test

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"testing"

	"github.com/lewisgibson/go-steam/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymmetricDecrypt_ValidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create some test data
	original := []byte("Hello, World!")

	// Act: encrypt the data
	encrypted, err := crypto.SymmetricEncrypt(original, key, nil)
	require.NoError(t, err)

	// Act: decrypt the data
	decrypted, err := crypto.SymmetricDecrypt(encrypted, key, false)
	require.NoError(t, err)

	// Assert: the decrypted data should be the same as the original data
	assert.Equal(t, original, decrypted)
}

func TestSymmetricDecrypt_WithHMAC(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create some test data
	original := []byte("Test with HMAC")

	// Act: encrypt the data using HMAC IV method
	encrypted, err := crypto.SymmetricEncryptWithHmacIv(original, key)
	require.NoError(t, err)

	// Act: decrypt the data with HMAC check
	decrypted, err := crypto.SymmetricDecrypt(encrypted, key, true)
	require.NoError(t, err)

	// Assert: the decrypted data should be the same as the original data
	assert.Equal(t, original, decrypted)
}

func TestSymmetricDecrypt_InvalidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	var testCases = []struct {
		name  string
		input []byte
		key   []byte
	}{
		{"Empty input", []byte{}, key},
		{"Short input", []byte("short"), key},
		{"Invalid key length", []byte("test"), []byte("invalid")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: decrypt the input
			_, err := crypto.SymmetricDecrypt(tc.input, tc.key, false)
			require.Error(t, err)
		})
	}
}

func TestSymmetricDecryptECB_ValidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create test data that's a multiple of block size
	original := bytes.Repeat([]byte("A"), 32)

	// Arrange: create the AES cipher
	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	// Act: apply PKCS7 padding manually
	padded := crypto.Pkcs7Pad(original, block.BlockSize())

	// Act: encrypt with ECB
	var encrypted = make([]byte, len(padded))
	for bs, be := 0, block.BlockSize(); bs < len(padded); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(encrypted[bs:be], padded[bs:be])
	}

	// Act: decrypt with our function
	decrypted, err := crypto.SymmetricDecryptECB(encrypted, key)
	require.NoError(t, err)

	// Assert: the decrypted data should be the same as the original data
	assert.Equal(t, original, decrypted)
}

func TestSymmetricDecryptECB_InvalidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	var testCases = []struct {
		name  string
		input []byte
		key   []byte
	}{
		{"Empty input", []byte{}, key},
		{"Invalid key length", []byte("test"), []byte("invalid")},
		{"Not multiple of block size", []byte("not multiple"), key},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: decrypt the input
			_, err := crypto.SymmetricDecryptECB(tc.input, tc.key)
			require.Error(t, err)
		})
	}
}
