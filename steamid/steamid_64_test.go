package steamid_test

import (
	"testing"

	"github.com/lbt05/go-steam/steamid"
	"github.com/stretchr/testify/require"
)

func TestFromSteam64_InvalidInput(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam64("invalid")
	require.Errorf(t, err, "expected error: %v", err)
	require.ErrorIsf(t, err, steamid.ErrInvalidSteam64, "unexpected error: %v", err)
	require.Emptyf(t, sid, "unexpected SteamID: %v", sid)
}

func TestFromSteam64_ValidInput(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam64("76561198065346589")
	require.NoErrorf(t, err, "unexpected error: %v", err)
	require.Equalf(t, steamid.SteamID(76561198065346589), sid, "unexpected SteamID: %v", sid)
}

func TestSteamID_ToSteam64(t *testing.T) {
	t.Parallel()

	sid := steamid.SteamID(76561198065346589)
	require.Equalf(t, "76561198065346589", sid.ToSteam64(), "unexpected Steam64: %v", sid.ToSteam64())
}
