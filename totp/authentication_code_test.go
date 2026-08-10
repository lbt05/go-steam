package totp_test

import (
	"testing"
	"time"

	"github.com/lbt05/go-steam/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAuthenticationCode_InvalidSecret(t *testing.T) {
	t.Parallel()

	// Act: attempt to generate an authentication code with an invalid shared secret.
	code, err := totp.CreateAuthenticationCode("not_a_valid_base64", time.Now())

	// Assert: an error is returned and the code is empty.
	require.Error(t, err, "expected an error for an invalid shared secret")
	assert.Empty(t, code, "expected empty code on error")
}

func TestCreateAuthenticationCode_Valid(t *testing.T) {
	t.Parallel()

	// Act: generate the authentication code for a valid shared secret and fixed time.
	code, err := totp.CreateAuthenticationCode("MTIzNDU2Nzg5MA==", time.Unix(0, 0))
	require.NoError(t, err, "should not error with a valid shared secret")
	assert.Len(t, code, 5, "authentication code should be 5 characters long")

	// Assert: the code contains only allowed characters.
	for _, r := range code {
		assert.Contains(t, "23456789BCDFGHJKMNPQRTVWXY", string(r), "code contains invalid character: %q", r)
	}
}

func TestCreateAuthenticationCode_Consistency(t *testing.T) {
	t.Parallel()

	// Act: generate the authentication code for a valid shared secret and fixed time twice.
	first, err := totp.CreateAuthenticationCode("MTIzNDU2Nzg5MA==", time.Unix(123456789, 0))
	require.NoErrorf(t, err, "error generating first code")
	require.NotEmptyf(t, first, "first code should not be empty")

	second, err := totp.CreateAuthenticationCode("MTIzNDU2Nzg5MA==", time.Unix(123456789, 0))
	require.NoErrorf(t, err, "error generating second code")
	require.NotEmptyf(t, second, "second code should not be empty")

	// Assert: the codes are identical.
	assert.Equal(t, first, second, "codes generated for the same input should be identical")
}

func TestCreateAuthenticationCode_DifferentTimes(t *testing.T) {
	t.Parallel()

	// Act: generate the authentication code for a valid shared secret at two different time steps.
	first, err := totp.CreateAuthenticationCode("MTIzNDU2Nzg5MA==", time.Unix(30, 0))
	require.NoError(t, err)
	require.NotEmptyf(t, first, "first code should not be empty")

	second, err := totp.CreateAuthenticationCode("MTIzNDU2Nzg5MA==", time.Unix(60, 0))
	require.NoError(t, err)
	require.NotEmptyf(t, second, "second code should not be empty")

	// Assert: the codes are different.
	assert.NotEqual(t, first, second, "codes generated at different time steps should differ")
}
