package iauthenticationservice

import (
	"context"
	"net/http"
)

// PollAuthSessionStatusParameters describes the parameters for the PollAuthSessionStatus method.
type PollAuthSessionStatusParameters struct {
	ClientID      uint64 `url:"client_id,omitempty"`
	RequestID     string `url:"request_id,omitempty"`
	TokenToRevoke uint64 `url:"token_to_revoke,omitempty"`
}

// PollAuthSessionStatusResponse describes the response for the PollAuthSessionStatus method.
type PollAuthSessionStatusResponse struct {
	NewClientID          string `json:"new_client_id"`
	NewChallengeURL      string `json:"new_challenge_url"`
	RefreshToken         string `json:"refresh_token"`
	AccessToken          string `json:"access_token"`
	HadRemoteInteraction bool   `json:"had_remote_interaction"`
	AccountName          string `json:"account_name"`
	NewGuardData         string `json:"new_guard_data"`
	AgreementSessionURL  string `json:"agreement_session_url"`
}

// PollAuthSessionStatus polls the status of an authentication session.
func (a *IAuthenticationService) PollAuthSessionStatus(ctx context.Context, params PollAuthSessionStatusParameters) (*PollAuthSessionStatusResponse, error) {
	var resBody struct {
		Response PollAuthSessionStatusResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodPost, "IAuthenticationService", "PollAuthSessionStatus", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
