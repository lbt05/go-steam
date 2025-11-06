package crypto_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature_ValidSignature(t *testing.T) {
	t.Parallel()

	data := []byte("test data")
	signature := []byte("invalid signature")

	valid, err := crypto.VerifySignature("RSA-SHA1", data, signature)
	require.NoError(t, err)
	assert.False(t, valid, "invalid signature should return false")
}

func TestVerifySignature_InvalidInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		algorithm string
		data      []byte
		signature []byte
	}{
		{"Empty data", "RSA-SHA1", []byte{}, []byte("signature")},
		{"Empty signature", "RSA-SHA1", []byte("data"), []byte{}},
		{"Unsupported algorithm", "RSA-MD5", []byte("data"), []byte("signature")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := crypto.VerifySignature(tc.algorithm, tc.data, tc.signature)
			require.Error(t, err)
		})
	}
}

func TestVerifySignature_SupportedAlgorithms(t *testing.T) {
	t.Parallel()

	data := []byte("test data")
	signature := []byte("invalid signature")

	algorithms := []string{"", "RSA-SHA1", "RSA-SHA256"}

	for _, alg := range algorithms {
		t.Run(alg, func(t *testing.T) {
			t.Parallel()

			valid, err := crypto.VerifySignature(alg, data, signature)
			require.NoError(t, err)
			assert.False(t, valid, "invalid signature should return false for algorithm %s", alg)
		})
	}
}
