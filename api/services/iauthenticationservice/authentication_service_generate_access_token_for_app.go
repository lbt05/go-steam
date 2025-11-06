package iauthenticationservice

import (
	"context"
	"net/http"

	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
)

// GenerateAccessTokenForAppParameters describes the parameters for the GenerateAccessTokenForApp method.
type GenerateAccessTokenForAppParameters struct {
	SteamID      steamid.SteamID         `url:"steamid,omitempty"`
	RefreshToken string                  `url:"refresh_token,omitempty"`
	RenewalType  steam.ETokenRenewalType `url:"renewal_type,omitempty"`
}

// GenerateAccessTokenForAppResponse describes the response for the GenerateAccessTokenForApp method.
type GenerateAccessTokenForAppResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// GenerateAccessTokenForApp generates an access token for a specific app.
func (a *IAuthenticationService) GenerateAccessTokenForApp(ctx context.Context, params GenerateAccessTokenForAppParameters) (*GenerateAccessTokenForAppResponse, error) {
	var resBody struct {
		Response GenerateAccessTokenForAppResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "IAuthenticationService", "GenerateAccessTokenForApp", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
