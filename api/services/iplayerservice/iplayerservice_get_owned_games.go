package iplayerservice

import (
	"context"
	"fmt"
	"net/http"
)

// GetOwnedGamesParameters describes the parameters for the GetOwnedGames method.
type GetOwnedGamesParameters struct {
	APIKey                 string   `url:"key,omitempty"`
	SteamID                uint64   `url:"steamid"`                             // SteamID of the player whose library to list.
	IncludeAppInfo         bool     `url:"include_appinfo,omitempty"`           // Include game name and icon URL in the response.
	IncludePlayedFreeGames bool     `url:"include_played_free_games,omitempty"` // Include free games the player has launched.
	AppIDsFilter           []uint32 `url:"appids_filter,omitempty"`             // Restrict results to the given app IDs.
	IncludeFreeSub         bool     `url:"include_free_sub,omitempty"`          // Include free-to-play games received as part of a free weekend or similar.
	Language               string   `url:"language,omitempty"`                  // Language for localized game names; defaults to English when empty.
}

// iconURLTemplate is the Steam CDN pattern for community game icons. Steam
// returns only the icon hash (img_icon_url); we assemble the full URL from
// the owning AppID and that hash so callers don't have to.
const iconURLTemplate = "https://media.steampowered.com/steamcommunity/public/images/apps/%d/%s.jpg"

// coverURLTemplate is the Steam CDN pattern for a game's library hero image
// (the wide artwork shown in the Steam library). The asset may not exist on
// the CDN for every game; callers should be prepared to handle 404s.
const coverURLTemplate = "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/%d/library_hero.jpg"

// logoURLTemplate is the Steam CDN pattern for a game's transparent logo PNG.
// The asset may not exist on the CDN for every game; callers should be
// prepared to handle 404s.
const logoURLTemplate = "https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/%d/logo.png"

// buildIconURL returns the canonical icon URL for a game given its AppID and
// the icon hash fragment returned by GetOwnedGames. Returns an empty string
// when the hash is empty, signalling that the game has no community icon.
func buildIconURL(appID uint32, hash string) string {
	if hash == "" {
		return ""
	}
	return fmt.Sprintf(iconURLTemplate, appID, hash)
}

// OwnedGame describes a single game in the user's owned library.
type OwnedGame struct {
	AppID                    uint32   `json:"appid"`
	Name                     string   `json:"name"`
	Playtime2Weeks           int32    `json:"playtime_2weeks"`
	PlaytimeForever          int32    `json:"playtime_forever"`
	GameIconHash             string   `json:"img_icon_url"`
	GameIconURL              string   `json:"-"`
	GameCoverURL             string   `json:"-"`
	GameLogoURL              string   `json:"-"`
	HasCommunityVisibleStats bool     `json:"has_community_visible_stats"`
	PlaytimeWindowsForever   int32    `json:"playtime_windows_forever"`
	PlaytimeMacForever       int32    `json:"playtime_mac_forever"`
	PlaytimeLinuxForever     int32    `json:"playtime_linux_forever"`
	RTimeLastPlayed          uint32   `json:"rtime_last_played"`
	CapsuleFilename          string   `json:"capsule_filename"`
	HasWorkshop              bool     `json:"has_workshop"`
	HasMarket                bool     `json:"has_market"`
	HasDLC                   bool     `json:"has_dlc"`
	HasLeaderboards          bool     `json:"has_leaderboards"`
	ContentDescriptorIDs     []uint32 `json:"content_descriptorids"`
}

// GetOwnedGamesResponse describes the response for the GetOwnedGames method.
type GetOwnedGamesResponse struct {
	GameCount uint32      `json:"game_count"`
	Games     []OwnedGame `json:"games"`
}

// GetOwnedGames returns the list of games owned by the given SteamID.
//
// This requires the Steam profile's game details to be set to "Public"; for
// private profiles Steam returns an empty list. Use IncludeAppInfo to also
// receive the localized game name and icon URL.
func (a *IPlayerService) GetOwnedGames(ctx context.Context, params GetOwnedGamesParameters) (*GetOwnedGamesResponse, error) {
	if params.Language == "" {
		params.Language = "en-US"
	}
	var resBody struct {
		Response GetOwnedGamesResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodGet, "IPlayerService", "GetOwnedGames", 1, params, &resBody); err != nil {
		return nil, err
	}
	for i := range resBody.Response.Games {
		g := &resBody.Response.Games[i]
		g.GameIconURL = buildIconURL(g.AppID, g.GameIconHash)
		g.GameCoverURL = fmt.Sprintf(coverURLTemplate, g.AppID)
		g.GameLogoURL = fmt.Sprintf(logoURLTemplate, g.AppID)
	}
	return &resBody.Response, nil
}
