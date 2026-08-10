package crypto_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/lbt05/go-steam/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymmetricEncrypt_ValidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create some test data
	input := []byte("Hello, World!")

	// Act: encrypt the data
	encrypted, err := crypto.SymmetricEncrypt(input, key, nil)
	require.NoError(t, err)

	// Assert: the encrypted data should be longer than the input due to padding and IV
	require.NotEmpty(t, encrypted)
	require.Greater(t, len(encrypted), len(input), "encrypted data should be longer than input due to padding and IV")
}

func TestSymmetricEncrypt_WithCustomIV(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create a random IV
	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	require.NoError(t, err)

	// Arrange: create some test data
	input := []byte("Test data")

	// Act: encrypt the data
	encrypted, err := crypto.SymmetricEncrypt(input, key, iv)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)
}

func TestSymmetricEncrypt_InvalidKey(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name string
		key  []byte
	}{
		{"Empty key", []byte{}},
		{"Short key", []byte("short")},
		{"Long key", make([]byte, 64)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: encrypt the data
			_, err := crypto.SymmetricEncrypt([]byte("test"), tc.key, nil)
			require.Error(t, err)

			// Assert: the error should contain the expected message
			assert.Contains(t, err.Error(), "key must be 32 bytes")
		})
	}
}

func TestSymmetricEncrypt_InvalidIV(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create an invalid IV
	invalidIV := []byte("invalid")

	// Act: encrypt the data
	_, err = crypto.SymmetricEncrypt([]byte("test"), key, invalidIV)
	require.Error(t, err)

	// Assert: the error should contain the expected message
	assert.Contains(t, err.Error(), "IV must be 16 bytes")
}

func TestSymmetricEncryptWithHmacIv_ValidInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create some test data
	input := []byte("Test data for HMAC IV")

	// Act: encrypt the data
	encrypted, err := crypto.SymmetricEncryptWithHmacIv(input, key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	// Assert: the encrypted data should be longer than the input due to padding and IV
	require.Greater(t, len(encrypted), len(input))
}

func TestSymmetricEncryptWithHmacIv_InvalidKey(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name string
		key  []byte
	}{
		{"Empty key", []byte{}},
		{"Short key", []byte("short")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: encrypt the data
			_, err := crypto.SymmetricEncryptWithHmacIv([]byte("test"), tc.key)
			require.Error(t, err)
		})
	}
}

func TestSymmetricEncryptWithHmacIv_EmptyInput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Act: encrypt the data
	_, err = crypto.SymmetricEncryptWithHmacIv([]byte{}, key)
	require.Error(t, err)

	// Assert: the error should contain the expected message
	assert.Contains(t, err.Error(), "input cannot be empty")
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	var testCases = []struct {
		name string
		data []byte
	}{
		{"Single byte", []byte("a")},
		{"Short data", []byte("hello")},
		{"Medium data", bytes.Repeat([]byte("A"), 100)},
		{"Long data", bytes.Repeat([]byte("B"), 1000)},
		{"Exact block size", bytes.Repeat([]byte("C"), 16)},
		{"Multiple block sizes", bytes.Repeat([]byte("D"), 32)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act: encrypt the data
			encrypted, err := crypto.SymmetricEncrypt(tc.data, key, nil)
			require.NoError(t, err)

			// Act: decrypt the data
			decrypted, err := crypto.SymmetricDecrypt(encrypted, key, false)
			require.NoError(t, err)

			// Assert: the decrypted data should be the same as the original data
			assert.Equal(t, tc.data, decrypted)
		})
	}
}

func TestDifferentIVsProduceDifferentOutput(t *testing.T) {
	t.Parallel()

	// Arrange: create a random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	// Arrange: create some test data
	data := []byte("same data")

	// Arrange: create two random IVs
	iv1 := make([]byte, 16)
	_, err = rand.Read(iv1)
	require.NoError(t, err)

	iv2 := make([]byte, 16)
	_, err = rand.Read(iv2)
	require.NoError(t, err)

	// Assert: the IVs are different
	require.NotEqual(t, iv1, iv2)

	// Act: encrypt the data
	encrypted1, err := crypto.SymmetricEncrypt(data, key, iv1)
	require.NoError(t, err)

	// Act: encrypt the data
	encrypted2, err := crypto.SymmetricEncrypt(data, key, iv2)
	require.NoError(t, err)

	// Assert: different IVs should produce different encrypted outputs
	assert.NotEqual(t, encrypted1, encrypted2)

	// Assert: both should decrypt to the same original data
	decrypted1, err := crypto.SymmetricDecrypt(encrypted1, key, false)
	require.NoError(t, err)

	decrypted2, err := crypto.SymmetricDecrypt(encrypted2, key, false)
	require.NoError(t, err)

	// Assert: the decrypted data should be the same as the original data
	assert.Equal(t, data, decrypted1)
	assert.Equal(t, data, decrypted2)
}
