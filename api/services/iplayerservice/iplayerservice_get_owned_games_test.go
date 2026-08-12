package iplayerservice_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/lbt05/go-steam/api/services/iplayerservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:embed fixtures/owned_games.json
var getOwnedGamesResponseBody []byte

func TestIPlayerService_GetOwnedGames_Success(t *testing.T) {
	t.Parallel()

	// Arrange: mock transport returns the embedded JSON body verbatim.
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), http.MethodGet, "IPlayerService", "GetOwnedGames", 1, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, verb, service, method string, version int, params, response any) error {
			return json.Unmarshal(getOwnedGamesResponseBody, response)
		})

	service := iplayerservice.NewIPlayerService(transport)

	// Act
	resp, err := service.GetOwnedGames(t.Context(), iplayerservice.GetOwnedGamesParameters{
		SteamID:        76561197960265728,
		IncludeAppInfo: true,
	})

	// Assert: shape decoded correctly
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint32(3), resp.GameCount)
	require.Len(t, resp.Games, 3)

	tf2 := resp.Games[0]
	require.Equal(t, uint32(440), tf2.AppID)
	require.Equal(t, "Team Fortress 2", tf2.Name)
	require.Equal(t, int32(12345), tf2.PlaytimeForever)
	require.Equal(t, int32(9000), tf2.PlaytimeWindowsForever)
	require.Equal(t, int32(3345), tf2.PlaytimeLinuxForever)
	require.True(t, tf2.HasWorkshop)
	require.True(t, tf2.HasDLC)
	require.False(t, tf2.HasLeaderboards)
	require.Equal(t, []uint32{1, 5}, tf2.ContentDescriptorIDs)
	require.Equal(t, "c5c8b2b6e8b1a3e4f7c9d0e1f2a3b4c5d6e7f8a9", tf2.GameIconHash)
	require.Equal(t, "https://media.steampowered.com/steamcommunity/public/images/apps/440/c5c8b2b6e8b1a3e4f7c9d0e1f2a3b4c5d6e7f8a9.jpg", tf2.GameIconURL)
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/440/library_hero.jpg", tf2.GameCoverURL)
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/440/logo.png", tf2.GameLogoURL)

	cs2 := resp.Games[1]
	require.Equal(t, uint32(730), cs2.AppID)
	require.Equal(t, "Counter-Strike 2", cs2.Name)
	require.Equal(t, uint32(1700123456), cs2.RTimeLastPlayed)
	require.False(t, cs2.HasDLC)
	require.Empty(t, cs2.ContentDescriptorIDs)
	require.Equal(t, "8b1d7c8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c", cs2.GameIconHash)
	require.Equal(t, "https://media.steampowered.com/steamcommunity/public/images/apps/730/8b1d7c8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c.jpg", cs2.GameIconURL)
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/730/library_hero.jpg", cs2.GameCoverURL)
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/730/logo.png", cs2.GameLogoURL)

	noIcon := resp.Games[2]
	require.Equal(t, uint32(10), noIcon.AppID)
	require.Empty(t, noIcon.GameIconHash, "missing img_icon_url should decode to empty hash")
	require.Empty(t, noIcon.GameIconURL, "missing icon hash should yield empty URL")
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/10/library_hero.jpg", noIcon.GameCoverURL, "cover URL is always populated from AppID")
	require.Equal(t, "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/10/logo.png", noIcon.GameLogoURL, "logo URL is always populated from AppID")
}

func TestIPlayerService_GetOwnedGames_DefaultsLanguageToEnUS(t *testing.T) {
	t.Parallel()

	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), http.MethodGet, "IPlayerService", "GetOwnedGames", 1, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, _ string, _ int, params, response any) error {
			p, ok := params.(iplayerservice.GetOwnedGamesParameters)
			require.True(t, ok)
			require.Equal(t, "en-US", p.Language)
			return json.Unmarshal(getOwnedGamesResponseBody, response)
		})

	service := iplayerservice.NewIPlayerService(transport)

	_, err := service.GetOwnedGames(t.Context(), iplayerservice.GetOwnedGamesParameters{
		SteamID: 1,
	})
	require.NoError(t, err)
}

func TestIPlayerService_GetOwnedGames_TransportError(t *testing.T) {
	t.Parallel()

	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("transport error"))

	service := iplayerservice.NewIPlayerService(transport)

	resp, err := service.GetOwnedGames(t.Context(), iplayerservice.GetOwnedGamesParameters{SteamID: 1})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestIPlayerService_NilTransport(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		iplayerservice.NewIPlayerService(nil)
	})
}