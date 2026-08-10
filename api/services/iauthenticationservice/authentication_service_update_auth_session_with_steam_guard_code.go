package iauthenticationservice

import (
	"context"
	"net/http"

	"github.com/lbt05/go-steam/language/steam"
	"github.com/lbt05/go-steam/steamid"
)

// UpdateAuthSessionWithSteamGuardCodeParameters describes the parameters for the UpdateAuthSessionWithSteamGuardCode method.
type UpdateAuthSessionWithSteamGuardCodeParameters struct {
	ClientID uint64                      `url:"client_id,omitempty"`
	SteamID  steamid.SteamID             `url:"steamid,omitempty"`
	Code     string                      `url:"code,omitempty"`
	CodeType steam.EAuthSessionGuardType `url:"code_type,omitempty"`
}

// UpdateAuthSessionWithSteamGuardCodeResponse describes the response for the UpdateAuthSessionWithSteamGuardCode method.
type UpdateAuthSessionWithSteamGuardCodeResponse struct{}

// UpdateAuthSessionWithSteamGuardCode updates an authentication session with Steam Guard code.
func (a *IAuthenticationService) UpdateAuthSessionWithSteamGuardCode(ctx context.Context, params UpdateAuthSessionWithSteamGuardCodeParameters) (*UpdateAuthSessionWithSteamGuardCodeResponse, error) {
	var resBody struct {
		Response UpdateAuthSessionWithSteamGuardCodeResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "IAuthenticationService", "UpdateAuthSessionWithSteamGuardCode", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
