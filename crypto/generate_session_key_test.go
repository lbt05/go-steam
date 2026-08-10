package crypto_test

import (
	"testing"

	"github.com/lbt05/go-steam/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSessionKey(t *testing.T) {
	t.Parallel()

	// Act: generate a session key
	plainKey, encryptedKey, err := crypto.GenerateSessionKey()
	require.NoError(t, err)

	// Assert: the plain and encrypted keys are as expected
	require.NotNil(t, plainKey)
	require.NotNil(t, encryptedKey)
	require.Len(t, plainKey, 32, "session key should be 32 bytes")
	require.NotEmpty(t, encryptedKey, "encrypted key should not be empty")
}

func TestGenerateSessionKey_Consistency(t *testing.T) {
	t.Parallel()

	// Act: generate multiple session keys
	var keys = make([][]byte, 10)
	for i := 0; i < 10; i++ {
		var err error
		keys[i], _, err = crypto.GenerateSessionKey()
		require.NoError(t, err)
	}

	// Assert: all keys are different
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			assert.NotEqual(t, keys[i], keys[j], "session keys should be unique")
		}
	}
}
