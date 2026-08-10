package steamid_test

import (
	"testing"

	"github.com/lbt05/go-steam/steamid"
	"github.com/stretchr/testify/assert"
)

func TestAccountID(t *testing.T) {
	t.Parallel()

	// Assert: the account id should be extracted from the steam id 64.
	assert.EqualValues(t, 105080861, steamid.SteamID(76561198065346589).AccountID())
}

func TestWithAccountID(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock steam id.
	original := steamid.SteamID(76561198065346589)

	// Act: modify the account id.
	modified := original.WithAccountID(12345678)

	// Assert: only the account id should be modified.
	assert.NotEqualValues(t, original.AccountID(), modified.AccountID())
	assert.EqualValues(t, original.AccountInstance(), modified.AccountInstance())
	assert.EqualValues(t, original.AccountType(), modified.AccountType())
	assert.EqualValues(t, original.AccountUniverse(), modified.AccountUniverse())

	// Assert: the account id should be modified.
	assert.EqualValues(t, 12345678, modified.AccountID())
}

func TestAccountInstance(t *testing.T) {
	t.Parallel()

	// Assert: the account instance should be extracted from the steam id 64.
	assert.EqualValues(t, 1, steamid.SteamID(76561198065346589).AccountInstance())
}

func TestWithAccountInstance(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock steam id.
	original := steamid.SteamID(76561198065346589)

	// Act: modify the account instance.
	modified := original.WithAccountInstance(steamid.AccountInstanceClanMask)

	// Assert: only the account instance should be modified.
	assert.EqualValues(t, original.AccountID(), modified.AccountID())
	assert.NotEqualValues(t, original.AccountInstance(), modified.AccountInstance())
	assert.EqualValues(t, original.AccountType(), modified.AccountType())
	assert.EqualValues(t, original.AccountUniverse(), modified.AccountUniverse())

	// Assert: the account instance should be modified.
	assert.EqualValues(t, steamid.AccountInstanceClanMask, modified.AccountInstance())
}

func TestAccountType(t *testing.T) {
	t.Parallel()

	// Assert: the account type should be extracted from the steam id 64.
	assert.EqualValues(t, steamid.AccountTypeIndividual, steamid.SteamID(76561198065346589).AccountType())
}

func TestWithAccountType(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock steam id.
	original := steamid.SteamID(76561198065346589)

	// Act: modify the account type.
	modified := original.WithAccountType(steamid.AccountTypeGameServer)

	// Assert: only the account type should be modified.
	assert.EqualValues(t, original.AccountID(), modified.AccountID())
	assert.EqualValues(t, original.AccountInstance(), modified.AccountInstance())
	assert.NotEqualValues(t, original.AccountType(), modified.AccountType())
	assert.EqualValues(t, original.AccountUniverse(), modified.AccountUniverse())

	// Assert: the account type should be modified.
	assert.EqualValues(t, steamid.AccountTypeGameServer, modified.AccountType())
}

func TestAccountUniverse(t *testing.T) {
	t.Parallel()

	// Assert: the account universe should be extracted from the steam id 64.
	assert.EqualValues(t, steamid.AccountUniversePublic, steamid.SteamID(76561198065346589).AccountUniverse())
}

func TestWithAccountUniverse(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock steam id.
	original := steamid.SteamID(76561198065346589)

	// Act: modify the account universe.
	modified := original.WithAccountUniverse(steamid.AccountUniverseBeta)

	// Assert: only the account universe should be modified.
	assert.EqualValues(t, original.AccountID(), modified.AccountID())
	assert.EqualValues(t, original.AccountInstance(), modified.AccountInstance())
	assert.EqualValues(t, original.AccountType(), modified.AccountType())
	assert.NotEqualValues(t, original.AccountUniverse(), modified.AccountUniverse())

	// Assert: the account universe should be modified.
	assert.EqualValues(t, steamid.AccountUniverseBeta, modified.AccountUniverse())
}
