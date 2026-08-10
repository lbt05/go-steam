package identity_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/api/services/itwofactorservice"
	"github.com/lbt05/go-steam/identity"
	"github.com/lbt05/go-steam/language/steam"
	"github.com/stretchr/testify/require"
)

// testRSAKey is shared across tests so RSA generation only runs once.
var (
	testRSAKeyOnce sync.Once
	testRSAKey     *rsa.PrivateKey
	testRSAKeyErr  error
)

func getTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	testRSAKeyOnce.Do(func() {
		testRSAKey, testRSAKeyErr = rsa.GenerateKey(rand.Reader, 2048)
	})
	require.NoError(t, testRSAKeyErr)
	return testRSAKey
}

// fakeResponse describes a single programmed response for fakeTransport.
type fakeResponse struct {
	method string // service method name to match; empty matches any
	body   any    // response body to populate via JSON round-trip
	err    error  // if non-nil, returned as the transport error
}

// fakeTransport is a programmable transports.Transport for tests. Responses
// are popped in order regardless of method.
type fakeTransport struct {
	mu        sync.Mutex
	responses []fakeResponse
	calls     []fakeCall
}

type fakeCall struct {
	service, method string
	params, response any
}

func newFakeTransport(responses ...fakeResponse) *fakeTransport {
	return &fakeTransport{responses: responses}
}

func (f *fakeTransport) Call(_ context.Context, _ /* verb */ string, service, method string, _ /* version */ int, params, response any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, fakeCall{service: service, method: method, params: params, response: response})

	if len(f.responses) == 0 {
		return errors.New("fakeTransport: no more programmed responses")
	}
	r := f.responses[0]
	f.responses = f.responses[1:]
	if r.err != nil {
		return r.err
	}
	if r.body == nil {
		return nil
	}
	// Steam's HTTP API wraps responses in {"response": ...}. Match that
	// envelope so callers' wrapper structs decode correctly.
	b, err := json.Marshal(map[string]any{"response": r.body})
	if err != nil {
		return err
	}
	return json.Unmarshal(b, response)
}

// fakeAPI wires the concrete IAuthenticationService and ITwoFactorService
// against a shared fakeTransport so tests can drive the full flow.
type fakeAPI struct {
	auth *iauthenticationservice.IAuthenticationService
	twof *itwofactorservice.ITwoFactorService
}

func newFakeAPI(t *testing.T, transport *fakeTransport) *fakeAPI {
	t.Helper()
	return &fakeAPI{
		auth: iauthenticationservice.NewIAuthenticationService(transport),
		twof: itwofactorservice.NewITwoFactorService(transport),
	}
}

func (f *fakeAPI) IAuthenticationService() *iauthenticationservice.IAuthenticationService { return f.auth }
func (f *fakeAPI) ITwoFactorService() *itwofactorservice.ITwoFactorService { return f.twof }

// rsaResponseFor returns the JSON body for a GetPasswordRSAPublicKey response
// containing the hex-encoded modulus and exponent of the test key plus a
// fixed timestamp.
func rsaResponseFor(t *testing.T, key *rsa.PrivateKey) map[string]string {
	t.Helper()
	return map[string]string{
		"publickey_mod": hex.EncodeToString(key.N.Bytes()),
		"publickey_exp": "10001",
		"timestamp":     "1700000000",
	}
}

// beginResponseFor builds a BeginAuthSessionViaCredentials response body with
// the given allowed confirmations. clientID and steamID are encoded as
// quoted strings to match the ,string JSON tags on the response struct.
func beginResponseFor(clientID uint64, requestID string, steamID uint64, interval int, allowed ...steam.EAuthSessionGuardType) map[string]any {
	confirmations := make([]map[string]any, len(allowed))
	for i, t := range allowed {
		confirmations[i] = map[string]any{
			"confirmation_type":  int32(t),
			"associated_message": "",
		}
	}
	return map[string]any{
		"client_id":             strconv.FormatUint(clientID, 10),
		"request_id":            requestID,
		"interval":              interval,
		"allowed_confirmations": confirmations,
		"steamid":               strconv.FormatUint(steamID, 10),
	}
}

// pollResponseFor builds a PollAuthSessionStatus response body. When
// refreshToken is non-empty the loop is expected to exit.
func pollResponseFor(refreshToken, accessToken string) map[string]string {
	return map[string]string{
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	}
}

func newTestProvider(t *testing.T, transport *fakeTransport, creds identity.Credentials) *identity.CredentialsIdentityProvider {
	t.Helper()
	api := newFakeAPI(t, transport)
	idp, err := identity.NewCredentialsIdentityProvider(creds, identity.WithAPI(api))
	require.NoError(t, err)
	return idp
}

func TestCredentialsIdentityProvider_BeginAuthSession_DeviceCodeAndEmailCodeAllowed(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(42, "req", 76561197960265728+1234, 5,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		)},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, uint64(42), session.ClientID)
	require.Equal(t, "req", session.RequestID)
	require.Equal(t, 5*time.Second, session.Interval)
	require.Len(t, session.AllowedConfirmations, 2)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode, session.AllowedConfirmations[0].ConfirmationType)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode, session.AllowedConfirmations[1].ConfirmationType)
	require.Equal(t, uint64(76561197960265728+1234), session.SteamID.Uint64())
}

func TestCredentialsIdentityProvider_BeginAuthSession_EmailCodeOnly_NoFailFast(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(7, "req", 76561197960265728+99, 4,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		)},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)
	require.Len(t, session.AllowedConfirmations, 1)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode, session.AllowedConfirmations[0].ConfirmationType)
}

func TestCredentialsIdentityProvider_BeginAuthSession_TransportError(t *testing.T) {
	t.Parallel()

	ft := newFakeTransport(fakeResponse{err: errors.New("boom")})

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.Error(t, err)
	require.Nil(t, session)
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_DeviceCode_PollsUntilToken(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(11, "req", 76561197960265728+1, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		)},
		// 1st poll: no token yet, loop continues
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("", "")},
		// 2nd poll: token issued
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("refresh-xyz", "access-xyz")},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	id, err := idp.SubmitSteamGuardCode(t.Context(), session, "12345")
	require.NoError(t, err)
	require.NotNil(t, id)
	require.Equal(t, "refresh-xyz", id.RefreshToken())
	require.Equal(t, "access-xyz", id.AccessToken())

	// Second call should hit the cache and not re-issue a Begin call.
	id2, err := idp.SubmitSteamGuardCode(t.Context(), session, "99999")
	require.NoError(t, err)
	require.Same(t, id, id2)
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_EmailCode_DoesNotCallTOTP(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(11, "req", 76561197960265728+1, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		)},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("r1", "a1")},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	id, err := idp.SubmitSteamGuardCode(t.Context(), session, "ABCDE")
	require.NoError(t, err)
	require.Equal(t, "r1", id.RefreshToken())

	// Verify UpdateAuthSessionWithSteamGuardCode was called with the email code.
	var updateCall *fakeCall
	for i := range ft.calls {
		if ft.calls[i].method == "UpdateAuthSessionWithSteamGuardCode" {
			updateCall = &ft.calls[i]
			break
		}
	}
	require.NotNil(t, updateCall)
	params, ok := updateCall.params.(iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters)
	require.True(t, ok)
	require.Equal(t, "ABCDE", params.Code)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode, params.CodeType)
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_NilSession(t *testing.T) {
	t.Parallel()

	idp := newTestProvider(t, newFakeTransport(), identity.Credentials{AccountName: "user", Password: "pass"})

	_, err := idp.SubmitSteamGuardCode(t.Context(), nil, "x")
	require.Error(t, err)
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_ContextCancelDuringPoll(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(1, "req", 76561197960265728+1, 60,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		)},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("", "")},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = idp.SubmitSteamGuardCode(ctx, session, "1")
	require.ErrorIs(t, err, context.Canceled)
}

func TestCredentialsIdentityProvider_Identity_SharedSecretShortcut(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(1, "req", 76561197960265728+7, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		)},
		fakeResponse{method: "QueryTime", body: map[string]string{"server_time": "1700000000"}},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("r-shared", "a-shared")},
	)

	// SharedSecret = "MTIzNDU2Nzg5MA==" (base64 of "1234567890") is the test
	// vector used in totp/authentication_code_test.go.
	idp := newTestProvider(t, ft, identity.Credentials{
		AccountName:  "user",
		Password:     "pass",
		SharedSecret: "MTIzNDU2Nzg5MA==",
	})

	id, err := idp.Identity(t.Context())
	require.NoError(t, err)
	require.Equal(t, "r-shared", id.RefreshToken())
	require.Equal(t, "a-shared", id.AccessToken())
}

func TestCredentialsIdentityProvider_Identity_RequiresTwoStepWithoutSharedSecret(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(1, "req", 76561197960265728+7, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		)},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	_, err := idp.Identity(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "shared secret is empty")
}

func TestCredentialsIdentityProvider_Identity_CachedReuse(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(1, "req", 76561197960265728+7, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		)},
		fakeResponse{method: "QueryTime", body: map[string]string{"server_time": "1700000000"}},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("r1", "a1")},
	)

	idp := newTestProvider(t, ft, identity.Credentials{
		AccountName:  "user",
		Password:     "pass",
		SharedSecret: "MTIzNDU2Nzg5MA==",
	})

	first, err := idp.Identity(t.Context())
	require.NoError(t, err)

	second, err := idp.Identity(t.Context())
	require.NoError(t, err)

	require.Same(t, first, second)
	require.Len(t, ft.calls, 5, "second Identity call must hit the cache without re-issuing RPCs")
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_PicksDeviceCodeWhenBothAllowed(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(11, "req", 76561197960265728+1, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode,
		)},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("r", "a")},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	_, err = idp.SubmitSteamGuardCode(t.Context(), session, "12345")
	require.NoError(t, err)

	var updateCall *fakeCall
	for i := range ft.calls {
		if ft.calls[i].method == "UpdateAuthSessionWithSteamGuardCode" {
			updateCall = &ft.calls[i]
			break
		}
	}
	require.NotNil(t, updateCall)
	params, ok := updateCall.params.(iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters)
	require.True(t, ok)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode, params.CodeType,
		"DeviceCode should win over EmailCode when both are allowed")
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_PicksEmailCodeWhenDeviceCodeAbsent(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(11, "req", 76561197960265728+1, 1,
			steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode,
		)},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: pollResponseFor("r", "a")},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	_, err = idp.SubmitSteamGuardCode(t.Context(), session, "ABCDE")
	require.NoError(t, err)

	var updateCall *fakeCall
	for i := range ft.calls {
		if ft.calls[i].method == "UpdateAuthSessionWithSteamGuardCode" {
			updateCall = &ft.calls[i]
			break
		}
	}
	require.NotNil(t, updateCall)
	params, ok := updateCall.params.(iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters)
	require.True(t, ok)
	require.Equal(t, steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode, params.CodeType)
}

func TestCredentialsIdentityProvider_SubmitSteamGuardCode_NoAllowedConfirmation(t *testing.T) {
	t.Parallel()

	key := getTestRSAKey(t)
	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: rsaResponseFor(t, key)},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: beginResponseFor(11, "req", 76561197960265728+1, 1)},
	)

	idp := newTestProvider(t, ft, identity.Credentials{AccountName: "user", Password: "pass"})

	session, err := idp.BeginAuthSession(t.Context())
	require.NoError(t, err)

	_, err = idp.SubmitSteamGuardCode(t.Context(), session, "12345")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no allowed confirmation type")
}