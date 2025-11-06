package steamid_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/lewisgibson/go-steam/steamid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromString_Steam2(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromString("STEAM_1:1:52540430")
	require.NoError(t, err)
	require.EqualValues(t, 0x11000010643681d, sid)
}

func TestFromString_Steam3(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromString("[U:1:52540430]")
	require.NoError(t, err)
	require.EqualValues(t, 0x11000010321b40e, sid)
}

func TestFromString_Steam64(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromString("76561198065346589")
	require.NoError(t, err)
	require.EqualValues(t, 76561198065346589, sid)
}

func TestFromString_Invalid(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromString("not a valid id")
	require.Error(t, err)
	require.Empty(t, sid)
}

func TestUint64(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, 76561198065346589, steamid.SteamID(76561198065346589).Uint64())
}

func TestString_Individual(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, "STEAM_1:1:52540430", steamid.SteamID(76561198065346589).String())
}

func TestString_NonIndividual(t *testing.T) {
	t.Parallel()

	// Arrange: create a clan steamid
	sid := steamid.NewSteamID(2469, 1, steamid.AccountTypeClan, steamid.AccountUniversePublic)

	// Assert: the string representation should be the same as the steam64 representation
	require.EqualValues(t, sid.ToSteam64(), sid.String())
}

func TestIsValid_Valid(t *testing.T) {
	t.Parallel()

	assert.True(t, steamid.SteamID(76561198065346589).IsValid())
}

func TestIsValid_Invalid(t *testing.T) {
	t.Parallel()

	// Assert: invalid account id
	assert.False(t, steamid.NewSteamID(0, 1, steamid.AccountTypeIndividual, steamid.AccountUniversePublic).IsValid())

	// Assert: invalid account universe
	assert.False(t, steamid.NewSteamID(1234, 1, steamid.AccountTypeIndividual, steamid.AccountUniverseInvalid).IsValid())

	// Assert: invalid account type
	assert.False(t, steamid.NewSteamID(1234, 1, steamid.AccountTypeInvalid, steamid.AccountUniversePublic).IsValid())
}

func TestConvertClanToChat(t *testing.T) {
	t.Parallel()

	input := steamid.NewSteamID(1234, 0, steamid.AccountTypeClan, steamid.AccountUniversePublic)
	expected := steamid.NewSteamID(1234, steamid.AccountInstanceClanMask, steamid.AccountTypeChat, steamid.AccountUniversePublic)
	require.EqualValues(t, expected, input.ConvertClanToChat())
}

func TestConvertChatToClan(t *testing.T) {
	t.Parallel()

	input := steamid.NewSteamID(1234, steamid.AccountInstanceClanMask, steamid.AccountTypeChat, steamid.AccountUniversePublic)
	expected := steamid.NewSteamID(1234, 0, steamid.AccountTypeClan, steamid.AccountUniversePublic)
	require.EqualValues(t, expected, input.ConvertChatToClan())
}

func TestFromUint64(t *testing.T) {
	t.Parallel()

	// Act: create from uint64
	sid := steamid.FromUint64(76561198065346589)

	// Assert: should create Steam ID with correct value
	require.EqualValues(t, 76561198065346589, sid)
}

func TestSteamID_EncodeValues(t *testing.T) {
	t.Parallel()

	// Arrange: create a valid Steam ID
	sid := steamid.SteamID(76561198065346589)
	values := &url.Values{}

	// Act: encode values
	require.NoError(t, sid.EncodeValues("steamid", values))

	// Assert: should succeed
	assert.Equal(t, "76561198065346589", values.Get("steamid"))
}

func TestSteamID_UnmarshalJSON_FromInteger(t *testing.T) {
	t.Parallel()

	// Act: unmarshal from JSON
	sid := steamid.SteamID(0)
	require.NoError(t, json.Unmarshal([]byte(`76561198065346589`), &sid))

	// Assert: should succeed
	assert.EqualValues(t, 76561198065346589, sid)
}

func TestSteamID_UnmarshalJSON_FromString(t *testing.T) {
	t.Parallel()

	// Act: unmarshal from JSON
	sid := steamid.SteamID(0)
	require.NoError(t, json.Unmarshal([]byte(`"76561198065346589"`), &sid))

	// Assert: should succeed
	assert.EqualValues(t, 76561198065346589, sid)
}

func TestSteamID_UnmarshalJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Act: unmarshal from JSON
	sid := steamid.SteamID(0)
	require.Error(t, json.Unmarshal([]byte(`{invalid json}`), &sid))
}
