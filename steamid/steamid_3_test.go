package steamid_test

import (
	"testing"

	"github.com/lbt05/go-steam/steamid"
	"github.com/stretchr/testify/require"
)

func TestFromSteam3_InvalidInput(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam3("invalid")
	require.Error(t, err)
	require.ErrorIs(t, err, steamid.ErrInvalidSteam3)
	require.EqualValues(t, 0, sid)
}

func TestFromSteam3_ValidInput(t *testing.T) {
	t.Parallel()

	sid, err := steamid.FromSteam3("[U:1:1234]")
	require.NoError(t, err)
	require.EqualValues(t, 0x1100001000004d2, sid)
}

func TestSteamID_ToSteam3(t *testing.T) {
	t.Parallel()

	require.Equal(t, "[U:1:1234]", steamid.NewSteamID(1234, 1, 1, steamid.AccountTypeIndividual).ToSteam3())
}
