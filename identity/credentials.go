package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/lewisgibson/go-steam/api"
	"github.com/lewisgibson/go-steam/api/services/iauthenticationservice"
	"github.com/lewisgibson/go-steam/api/services/itwofactorservice"
	"github.com/lewisgibson/go-steam/api/transports"
	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
	"github.com/lewisgibson/go-steam/totp"
)

// Identity represents access for a Steam account.
type Identity struct {
	steamID      steamid.SteamID
	accessToken  string
	refreshToken string
	createdAt    time.Time
	expiresAt    time.Time
}

// SteamID returns the SteamID of the identity.
func (i *Identity) SteamID() steamid.SteamID {
	return i.steamID
}

// AccessToken returns the access token of the identity.
func (i *Identity) AccessToken() string {
	return i.accessToken
}

// RefreshToken returns the refresh token of the identity.
func (i *Identity) RefreshToken() string {
	return i.refreshToken
}

// CreatedAt returns the created at time of the identity.
func (i *Identity) CreatedAt() time.Time {
	return i.createdAt
}

// ExpiresAt returns the expires at time of the identity.
func (i *Identity) ExpiresAt() time.Time {
	return i.expiresAt
}

// Credentials identify a single Steam account.
type Credentials struct {
	AccountName  string
	Password     string
	SharedSecret string
}

// CredentialsIdentityProvider is the identity provider.
type CredentialsIdentityProvider struct {
	api         API
	credentials Credentials

	mutex    sync.Mutex
	identity *Identity
}

// CredentialsIdentityProviderOption is a function that can be used to configure the identity provider.
type CredentialsIdentityProviderOption func(*CredentialsIdentityProvider)

// API are the services of the identity provider.
type API interface {
	IAuthenticationService() *iauthenticationservice.IAuthenticationService
	ITwoFactorService() *itwofactorservice.ITwoFactorService
}

// WithAPI sets the API for the identity provider.
func WithAPI(api API) CredentialsIdentityProviderOption {
	return func(idp *CredentialsIdentityProvider) {
		idp.api = api
	}
}

// NewCredentialsIdentityProvider creates a new identity provider using credentials.
func NewCredentialsIdentityProvider(credentials Credentials, opts ...CredentialsIdentityProviderOption) (*CredentialsIdentityProvider, error) {
	var idp = &CredentialsIdentityProvider{
		credentials: credentials,
	}
	for _, opt := range opts {
		opt(idp)
	}
	if idp.api == nil {
		api, err := api.NewAPI(transports.NewHTTPTransport())
		if err != nil {
			return nil, fmt.Errorf("failed to create API: %w", err)
		}
		idp.api = api
	}
	return idp, nil
}

// Identity returns the identity of the client.
func (idp *CredentialsIdentityProvider) Identity(ctx context.Context) (*Identity, error) {
	idp.mutex.Lock()
	defer idp.mutex.Unlock()

	if idp.identity == nil || idp.identity.expiresAt.Before(time.Now()) {
		encryptedPassword, timestamp, err := idp.encryptPassword(ctx, idp.credentials.AccountName, idp.credentials.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}

		authSession, err := idp.beginAuthSessionViaCredentials(ctx, idp.credentials.AccountName, encryptedPassword, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to begin auth session: %w", err)
		}

		var found bool
		for _, confirmation := range authSession.AllowedConfirmations {
			if confirmation.ConfirmationType == steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("no allowed confirmation found")
		}

		var authSessionStatus *iauthenticationservice.PollAuthSessionStatusResponse
		for {
			select {
			default:
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			authSessionStatus, err = idp.api.IAuthenticationService().PollAuthSessionStatus(ctx, iauthenticationservice.PollAuthSessionStatusParameters{
				ClientID:  authSession.ClientID,
				RequestID: authSession.RequestID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to poll auth session: %w", err)
			}

			if authSessionStatus.RefreshToken != "" {
				break
			}

			if authSessionStatus.NewChallengeURL == "" {
				serverTimeResp, err := idp.api.ITwoFactorService().QueryTime(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get Steam server time: %w", err)
				}

				tfa, err := totp.CreateAuthenticationCode(idp.credentials.SharedSecret, time.Unix(serverTimeResp.ServerTime, 0))
				if err != nil {
					return nil, fmt.Errorf("failed to generate TFA code: %w", err)
				}

				_, err = idp.api.IAuthenticationService().UpdateAuthSessionWithSteamGuardCode(ctx, iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters{
					ClientID: authSession.ClientID,
					SteamID:  authSession.SteamID,
					Code:     tfa,
					CodeType: steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to update with 2FA: %w", err)
				}
			}

			select {
			case <-time.After(time.Duration(authSession.Interval) * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		idp.identity = &Identity{
			steamID:      authSession.SteamID,
			accessToken:  authSessionStatus.AccessToken,
			refreshToken: authSessionStatus.RefreshToken,
			createdAt:    time.Now(),
			expiresAt:    time.Now().Add(time.Hour), // This is not documented anywhere, but it seems to be the case.
		}
	}

	return idp.identity, nil
}

// encryptPassword encrypts the password using the RSA public key.
func (idp *CredentialsIdentityProvider) encryptPassword(ctx context.Context, accountName, password string) (string, uint64, error) {
	passwordRSAPubKeyResp, err := idp.api.IAuthenticationService().GetPasswordRSAPublicKey(ctx, iauthenticationservice.GetPasswordRSAPublicKeyParameters{
		AccountName: accountName,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to get RSA public key: %w", err)
	}

	modulus, ok := new(big.Int).SetString(passwordRSAPubKeyResp.PublicKeyMod, 16)
	if !ok {
		return "", 0, fmt.Errorf("failed to parse modulus as hex: %s", passwordRSAPubKeyResp.PublicKeyMod)
	}

	exp, err := strconv.ParseInt(passwordRSAPubKeyResp.PublicKeyExp, 16, 64)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse exponent as hex: %s", passwordRSAPubKeyResp.PublicKeyExp)
	}

	key := &rsa.PublicKey{
		N: modulus,
		E: int(exp),
	}

	encryptedPasswordBytes, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(password))
	if err != nil {
		return "", 0, fmt.Errorf("failed to encrypt password: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encryptedPasswordBytes), passwordRSAPubKeyResp.Timestamp, nil
}

// beginAuthSessionViaCredentials begins an auth session via credentials.
func (idp *CredentialsIdentityProvider) beginAuthSessionViaCredentials(ctx context.Context, accountName, encryptedPassword string, encryptionTimestamp uint64) (*iauthenticationservice.BeginAuthSessionViaCredentialsResponse, error) {
	return idp.api.IAuthenticationService().BeginAuthSessionViaCredentials(ctx, iauthenticationservice.BeginAuthSessionViaCredentialsParameters{
		AccountName:         accountName,
		EncryptedPassword:   encryptedPassword,
		EncryptionTimestamp: encryptionTimestamp,
		PlatformType:        int32(steam.EAuthTokenPlatformType_k_EAuthTokenPlatformType_SteamClient),
		Persistence:         int32(steam.ESessionPersistence_k_ESessionPersistence_Persistent),
		WebsiteID:           "Community",
		DeviceDetails: iauthenticationservice.BeginAuthSessionViaCredentialsDeviceDetailsParameters{
			DeviceFriendlyName: "go-steam",
			PlatformType:       int32(steam.EAuthTokenPlatformType_k_EAuthTokenPlatformType_SteamClient),
		},
		QosLevel: 4,
	})
}
