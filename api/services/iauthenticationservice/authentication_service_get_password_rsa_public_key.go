package iauthenticationservice

import (
	"context"
	"net/http"
)

// GetPasswordRSAPublicKeyParameters describes the parameters for the GetPasswordRSAPublicKey method.
type GetPasswordRSAPublicKeyParameters struct {
	AccountName string `url:"account_name,omitempty"`
}

// GetPasswordRSAPublicKeyResponse describes the response for the GetPasswordRSAPublicKey method.
type GetPasswordRSAPublicKeyResponse struct {
	PublicKeyMod string `json:"publickey_mod"`
	PublicKeyExp string `json:"publickey_exp"`
	Timestamp    uint64 `json:"timestamp,string"`
}

// GetPasswordRSAPublicKey gets the RSA public key for password encryption.
func (a *IAuthenticationService) GetPasswordRSAPublicKey(ctx context.Context, params GetPasswordRSAPublicKeyParameters) (*GetPasswordRSAPublicKeyResponse, error) {
	var resBody struct {
		Response GetPasswordRSAPublicKeyResponse `json:"response"`
	}
	if err := a.transport.Call(ctx, http.MethodGet, "IAuthenticationService", "GetPasswordRSAPublicKey", 1, params, &resBody); err != nil {
		return nil, err
	}
	return &resBody.Response, nil
}
