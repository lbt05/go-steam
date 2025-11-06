package crypto_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPkcs7Pad(t *testing.T) {
	t.Parallel()

	data := []byte("test data")
	pad := crypto.Pkcs7Pad(data, 16)
	assert.Equal(t, []byte("test data\x07\x07\x07\x07\x07\x07\x07"), pad)
}

func TestPkcs7Unpad(t *testing.T) {
	t.Parallel()

	data := []byte("test data\x07\x07\x07\x07\x07\x07\x07")
	unpad, err := crypto.Pkcs7Unpad(data, 16)
	require.NoError(t, err)
	assert.Equal(t, []byte("test data"), unpad)
}
