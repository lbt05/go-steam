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

	"github.com/lbt05/go-steam/api"
	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/api/services/itwofactorservice"
	"github.com/lbt05/go-steam/api/transports"
	"github.com/lbt05/go-steam/language/steam"
	"github.com/lbt05/go-steam/steamid"
	"github.com/lbt05/go-steam/totp"
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

// AuthSession represents an in-progress Steam authentication session returned
// by BeginAuthSession. The session must be passed to SubmitSteamGuardCode
// together with the user-supplied Steam Guard code to complete login.
type AuthSession struct {
	ClientID             uint64
	RequestID            string
	SteamID              steamid.SteamID
	Interval             time.Duration
	AllowedConfirmations []iauthenticationservice.AllowedConfirmation
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

// BeginAuthSession performs the first half of the login flow: it encrypts the
// password with Steam's RSA public key and opens an authentication session,
// returning the session state required by SubmitSteamGuardCode.
//
// The returned AuthSession exposes AllowedConfirmations so callers can decide
// which Steam Guard code type (EmailCode or DeviceCode) to submit next.
func (idp *CredentialsIdentityProvider) BeginAuthSession(ctx context.Context) (*AuthSession, error) {
	idp.mutex.Lock()
	defer idp.mutex.Unlock()

	encryptedPassword, timestamp, err := idp.encryptPassword(ctx, idp.credentials.AccountName, idp.credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	authSession, err := idp.beginAuthSessionViaCredentials(ctx, idp.credentials.AccountName, encryptedPassword, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to begin auth session: %w", err)
	}

	return &AuthSession{
		ClientID:             authSession.ClientID,
		RequestID:            authSession.RequestID,
		SteamID:              authSession.SteamID,
		Interval:             time.Duration(authSession.Interval) * time.Second,
		AllowedConfirmations: authSession.AllowedConfirmations,
	}, nil
}

// SubmitSteamGuardCode performs the second half of the login flow: it submits
// the user-supplied Steam Guard code to Steam, polls until a refresh token is
// issued, and caches the resulting identity for subsequent calls.
//
// The code type (e.g. DeviceCode for TOTP or EmailCode for email 2FA) is
// chosen automatically from session.AllowedConfirmations with the following
// priority order: DeviceCode, EmailCode, DeviceConfirmation, EmailConfirmation,
// MachineToken. Unknown, None and LegacyMachineAuth are ignored.
func (idp *CredentialsIdentityProvider) SubmitSteamGuardCode(ctx context.Context, session *AuthSession, code string) (*Identity, error) {
	if session == nil {
		return nil, fmt.Errorf("auth session is nil")
	}

	codeType, ok := pickAllowedConfirmationType(session.AllowedConfirmations)
	if !ok {
		return nil, fmt.Errorf("no allowed confirmation type can accept a code")
	}

	idp.mutex.Lock()
	defer idp.mutex.Unlock()

	if idp.identity != nil && idp.identity.expiresAt.After(time.Now()) {
		return idp.identity, nil
	}

	_, err := idp.api.IAuthenticationService().UpdateAuthSessionWithSteamGuardCode(ctx, iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters{
		ClientID: session.ClientID,
		SteamID:  session.SteamID,
		Code:     code,
		CodeType: codeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update with Steam Guard code: %w", err)
	}

	var authSessionStatus *iauthenticationservice.PollAuthSessionStatusResponse
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		authSessionStatus, err = idp.api.IAuthenticationService().PollAuthSessionStatus(ctx, iauthenticationservice.PollAuthSessionStatusParameters{
			ClientID:  session.ClientID,
			RequestID: session.RequestID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to poll auth session: %w", err)
		}

		if authSessionStatus.RefreshToken != "" {
			break
		}

		select {
		case <-time.After(session.Interval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	idp.identity = &Identity{
		steamID:      session.SteamID,
		accessToken:  authSessionStatus.AccessToken,
		refreshToken: authSessionStatus.RefreshToken,
		createdAt:    time.Now(),
		expiresAt:    time.Now().Add(time.Hour), // This is not documented anywhere, but it seems to be the case.
	}

	return idp.identity, nil
}

// Identity runs the full single-shot login flow when SharedSecret is available:
// it generates a TOTP code from the shared secret and submits it as a
// DeviceCode Steam Guard code. If SharedSecret is empty, callers should use
// BeginAuthSession and SubmitSteamGuardCode instead.
func (idp *CredentialsIdentityProvider) Identity(ctx context.Context) (*Identity, error) {
	idp.mutex.Lock()
	defer idp.mutex.Unlock()

	if idp.identity != nil && idp.identity.expiresAt.After(time.Now()) {
		return idp.identity, nil
	}

	session, err := idp.beginSessionLocked(ctx)
	if err != nil {
		return nil, err
	}

	if idp.credentials.SharedSecret == "" {
		return nil, fmt.Errorf("shared secret is empty; call BeginAuthSession and SubmitSteamGuardCode to provide a Steam Guard code")
	}

	serverTimeResp, err := idp.api.ITwoFactorService().QueryTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Steam server time: %w", err)
	}

	tfa, err := totp.CreateAuthenticationCode(idp.credentials.SharedSecret, time.Unix(serverTimeResp.ServerTime, 0))
	if err != nil {
		return nil, fmt.Errorf("failed to generate TFA code: %w", err)
	}

	_, err = idp.api.IAuthenticationService().UpdateAuthSessionWithSteamGuardCode(ctx, iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters{
		ClientID: session.ClientID,
		SteamID:  session.SteamID,
		Code:     tfa,
		CodeType: steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update with 2FA: %w", err)
	}

	var authSessionStatus *iauthenticationservice.PollAuthSessionStatusResponse
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		authSessionStatus, err = idp.api.IAuthenticationService().PollAuthSessionStatus(ctx, iauthenticationservice.PollAuthSessionStatusParameters{
			ClientID:  session.ClientID,
			RequestID: session.RequestID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to poll auth session: %w", err)
		}

		if authSessionStatus.RefreshToken != "" {
			break
		}

		select {
		case <-time.After(session.Interval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	idp.identity = &Identity{
		steamID:      session.SteamID,
		accessToken:  authSessionStatus.AccessToken,
		refreshToken: authSessionStatus.RefreshToken,
		createdAt:    time.Now(),
		expiresAt:    time.Now().Add(time.Hour),
	}

	return idp.identity, nil
}

// beginSessionLocked performs the BeginAuthSession flow assuming idp.mutex is
// already held by the caller.
func (idp *CredentialsIdentityProvider) beginSessionLocked(ctx context.Context) (*AuthSession, error) {
	encryptedPassword, timestamp, err := idp.encryptPassword(ctx, idp.credentials.AccountName, idp.credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	authSession, err := idp.beginAuthSessionViaCredentials(ctx, idp.credentials.AccountName, encryptedPassword, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to begin auth session: %w", err)
	}

	return &AuthSession{
		ClientID:             authSession.ClientID,
		RequestID:            authSession.RequestID,
		SteamID:              authSession.SteamID,
		Interval:             time.Duration(authSession.Interval) * time.Second,
		AllowedConfirmations: authSession.AllowedConfirmations,
	}, nil
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

// pickAllowedConfirmationType selects the EAuthSessionGuardType that should be
// used to submit a user-supplied Steam Guard code from the confirmations Steam
// has authorised for this session. Higher-priority types are returned first:
//
//	DeviceCode          — TOTP, the common automated path.
//	EmailCode           — email 2FA code, the common interactive path.
//	DeviceConfirmation  — mobile-app push approval.
//	EmailConfirmation   — email link approval.
//	MachineToken        — pre-saved machine token.
//
// Unknown, None and LegacyMachineAuth are deliberately ignored. The second
// return value is false when none of the supported types is allowed.
func pickAllowedConfirmationType(confirmations []iauthenticationservice.AllowedConfirmation) (steam.EAuthSessionGuardType, bool) {
	priority := []steam.EAuthSessionGuardType{
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceConfirmation,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailConfirmation,
		steam.EAuthSessionGuardType_k_EAuthSessionGuardType_MachineToken,
	}
	allowed := make(map[steam.EAuthSessionGuardType]struct{}, len(confirmations))
	for _, c := range confirmations {
		allowed[c.ConfirmationType] = struct{}{}
	}
	for _, p := range priority {
		if _, ok := allowed[p]; ok {
			return p, true
		}
	}
	return 0, false
}