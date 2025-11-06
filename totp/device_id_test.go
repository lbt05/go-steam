package totp_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/totp"
	"github.com/stretchr/testify/assert"
)

func TestCreateDeviceID_EmptySteamID(t *testing.T) {
	t.Parallel()

	// Assert: The SHA-1 checksum for an empty string is 0xda39a3ee5e6b4b0d3255bfef95601890afd80709
	assert.Equal(t, "android:da39a3ee-5e6b-4b0d-3255-bfef95601890", totp.CreateDeviceID(""))
}

func TestCreateDeviceID_ValidSteamID(t *testing.T) {
	t.Parallel()

	// Assert: The SHA-1 checksum for "76561198065346589" is 0xa5641b97119c58861b798a30dd02ef6e
	assert.Equal(t, "android:a5641b97-119c-5886-1b79-8a30dd02ef6e", totp.CreateDeviceID("76561198065346589"))
}
