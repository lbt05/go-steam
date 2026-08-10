package steamid_test

import (
	"testing"

	"github.com/lbt05/go-steam/steamid"
	"github.com/stretchr/testify/require"
)

func TestFromSteam2_InvalidInput(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam2("invalid")
	require.Error(t, err)
	require.ErrorIs(t, err, steamid.ErrInvalidSteam2)
	require.EqualValues(t, 0, sid)
}

func TestFromSteam2_ValidInput_Steam0(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam2("STEAM_0:1:1234")
	require.NoError(t, err)
	require.EqualValues(t, 0x1100001000009a5, sid)
}

func TestFromSteam2_ValidInput_Steam2(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam2("STEAM_2:0:9999")
	require.NoError(t, err)
	require.EqualValues(t, 0x120000100004e1e, sid)
}

func TestToSteam2_ValidUniverse(t *testing.T) {
	t.Parallel()

	require.EqualValues(t, "STEAM_1:0:9999", steamid.NewSteamID(19998, 1, 2, steamid.AccountTypeIndividual).ToSteam2())
}

func TestToSteam2_InvalidUniverse(t *testing.T) {
	t.Parallel()

	// Assert: UniverseInvalid (0) should be treated as UniversePublic (1).
	require.EqualValues(t, "STEAM_1:1:1234", steamid.NewSteamID(2469, 1, steamid.AccountUniverseInvalid, steamid.AccountTypeIndividual).ToSteam2())
}
