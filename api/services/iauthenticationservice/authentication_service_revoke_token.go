package iauthenticationservice

import (
	"context"
	"net/http"

	"github.com/lbt05/go-steam/language/steam"
)

// RevokeTokenParameters describes the parameters for the RevokeToken method.
type RevokeTokenParameters struct {
	Token        string                       `url:"token,omitempty"`
	RevokeAction steam.EAuthTokenRevokeAction `url:"revoke_action,omitempty"`
}

// RevokeTokenResponse describes the response for the RevokeToken method.
type RevokeTokenResponse struct{}

// RevokeToken revokes an access or refresh token.
func (a *IAuthenticationService) RevokeToken(ctx context.Context, params RevokeTokenParameters) (*RevokeTokenResponse, error) {
	var resBody struct {
		Response RevokeTokenResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "IAuthenticationService", "RevokeToken", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
