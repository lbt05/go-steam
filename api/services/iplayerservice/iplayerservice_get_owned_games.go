package iplayerservice

import (
	"context"
	"net/http"
)

// GetOwnedGamesParameters describes the parameters for the GetOwnedGames method.
type GetOwnedGamesParameters struct {
	SteamID                uint64   `url:"steamid"`                            // SteamID of the player whose library to list.
	IncludeAppInfo         bool     `url:"include_appinfo,omitempty"`           // Include game name and icon URL in the response.
	IncludePlayedFreeGames bool     `url:"include_played_free_games,omitempty"` // Include free games the player has launched.
	AppIDsFilter           []uint32 `url:"appids_filter,omitempty"`            // Restrict results to the given app IDs.
	IncludeFreeSub         bool     `url:"include_free_sub,omitempty"`         // Include free-to-play games received as part of a free weekend or similar.
	Language               string   `url:"language,omitempty"`                 // Language for localized game names; defaults to English when empty.
}

// OwnedGame describes a single game in the user's owned library.
type OwnedGame struct {
	AppID                     uint32   `json:"appid"`
	Name                      string   `json:"name"`
	Playtime2Weeks            int32    `json:"playtime_2weeks"`
	PlaytimeForever           int32    `json:"playtime_forever"`
	ImgIconURL                string   `json:"img_icon_url"`
	HasCommunityVisibleStats  bool     `json:"has_community_visible_stats"`
	PlaytimeWindowsForever    int32    `json:"playtime_windows_forever"`
	PlaytimeMacForever        int32    `json:"playtime_mac_forever"`
	PlaytimeLinuxForever      int32    `json:"playtime_linux_forever"`
	RTimeLastPlayed           uint32   `json:"rtime_last_played"`
	CapsuleFilename           string   `json:"capsule_filename"`
	HasWorkshop               bool     `json:"has_workshop"`
	HasMarket                 bool     `json:"has_market"`
	HasDLC                    bool     `json:"has_dlc"`
	HasLeaderboards           bool     `json:"has_leaderboards"`
	ContentDescriptorIDs      []uint32 `json:"content_descriptorids"`
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
	return &resBody.Response, nil
}