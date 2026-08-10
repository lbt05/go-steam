package client_test

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

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/api/services/itwofactorservice"
	"github.com/lbt05/go-steam/client"
	"github.com/lbt05/go-steam/identity"
	"github.com/lbt05/go-steam/language/steam"
	"github.com/lbt05/go-steam/protocol"
	"github.com/stretchr/testify/require"
)

// fakeAPI is a minimal identity.API implementation backed by a shared
// fakeTransport. The SteamDirectory accessor returns nil because these tests
// do not exercise Client.Connect.
type fakeAPI struct {
	auth *iauthenticationservice.IAuthenticationService
	twof *itwofactorservice.ITwoFactorService
}

func (f *fakeAPI) IAuthenticationService() *iauthenticationservice.IAuthenticationService { return f.auth }
func (f *fakeAPI) ITwoFactorService() *itwofactorservice.ITwoFactorService { return f.twof }

// fakeResponse describes one programmed response for fakeTransport.
type fakeResponse struct {
	method string
	body   any
	err    error
}

type fakeTransport struct {
	mu        sync.Mutex
	responses []fakeResponse
}

func newFakeTransport(responses ...fakeResponse) *fakeTransport {
	return &fakeTransport{responses: responses}
}

func (f *fakeTransport) Call(_ context.Context, _ /* verb */ string, service, method string, _ /* version */ int, _, response any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
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
	b, err := json.Marshal(map[string]any{"response": r.body})
	if err != nil {
		return err
	}
	return json.Unmarshal(b, response)
}

func newClientWithFakeAPI(t *testing.T, ft *fakeTransport) *client.Client {
	t.Helper()
	api := &fakeAPI{
		auth: iauthenticationservice.NewIAuthenticationService(ft),
		twof: itwofactorservice.NewITwoFactorService(ft),
	}
	c, err := client.NewClient(client.WithAuthAPI(api))
	require.NoError(t, err)
	return c
}

func TestClient_BeginAuthSession_EmitsEventSteamGuardChallenge(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: map[string]string{
			"publickey_mod": hex.EncodeToString(key.N.Bytes()),
			"publickey_exp": "10001",
			"timestamp":     "1700000000",
		}},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: map[string]any{
			"client_id":  "1",
			"request_id": "req",
			"interval":   5,
			"allowed_confirmations": []map[string]any{
				{"confirmation_type": int32(steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode), "associated_message": ""},
				{"confirmation_type": int32(steam.EAuthSessionGuardType_k_EAuthSessionGuardType_DeviceCode), "associated_message": ""},
			},
			"steamid": strconv.FormatUint(76561197960265728+1, 10),
		}},
	)

	c := newClientWithFakeAPI(t, ft)

	session, err := c.BeginAuthSession(t.Context(), &client.Credentials{
		Username: "user",
		Password: "pass",
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	select {
	case event := <-c.Events():
		challenge, ok := event.(*protocol.EventSteamGuardChallenge)
		require.True(t, ok, "expected EventSteamGuardChallenge, got %T", event)
		require.Same(t, session, challenge.Session)
		require.Len(t, challenge.AllowedConfirmations, 2)
	default:
		t.Fatal("expected EventSteamGuardChallenge on events channel")
	}
}

func TestClient_SubmitSteamGuardCode_AfterBeginAuthSession(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: map[string]string{
			"publickey_mod": hex.EncodeToString(key.N.Bytes()),
			"publickey_exp": "10001",
			"timestamp":     "1700000000",
		}},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: map[string]any{
			"client_id":  "2",
			"request_id": "req",
			"interval":   1,
			"allowed_confirmations": []map[string]any{
				{"confirmation_type": int32(steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode), "associated_message": ""},
			},
			"steamid": strconv.FormatUint(76561197960265728+2, 10),
		}},
		fakeResponse{method: "UpdateAuthSessionWithSteamGuardCode"},
		fakeResponse{method: "PollAuthSessionStatus", body: map[string]string{
			"refresh_token": "refresh-abc",
			"access_token":  "access-abc",
		}},
	)

	c := newClientWithFakeAPI(t, ft)

	session, err := c.BeginAuthSession(t.Context(), &client.Credentials{
		Username: "user",
		Password: "pass",
	})
	require.NoError(t, err)

	// Drain the challenge event so we can verify the next event channel state.
	<-c.Events()

	id, err := c.SubmitSteamGuardCode(t.Context(), session, "ABCDE")
	require.NoError(t, err)
	require.NotNil(t, id)

	cached, err := c.Identity(t.Context())
	require.NoError(t, err)
	require.Equal(t, id.RefreshToken(), cached.RefreshToken())
	require.Equal(t, "refresh-abc", cached.RefreshToken())
}

func TestClient_Identity_BeforeAuth(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ft := newFakeTransport(
		fakeResponse{method: "GetPasswordRSAPublicKey", body: map[string]string{
			"publickey_mod": hex.EncodeToString(key.N.Bytes()),
			"publickey_exp": "10001",
			"timestamp":     "1700000000",
		}},
		fakeResponse{method: "BeginAuthSessionViaCredentials", body: map[string]any{
			"client_id":  "3",
			"request_id": "req",
			"interval":   1,
			"allowed_confirmations": []map[string]any{
				{"confirmation_type": int32(steam.EAuthSessionGuardType_k_EAuthSessionGuardType_EmailCode), "associated_message": ""},
			},
			"steamid": strconv.FormatUint(76561197960265728+3, 10),
		}},
	)

	c := newClientWithFakeAPI(t, ft)

	_, err = c.Identity(t.Context())
	require.Error(t, err)
	require.Contains(t, err.Error(), "identity not set")
}

func TestClient_BeginAuthSession_NilCredentials(t *testing.T) {
	t.Parallel()

	c, err := client.NewClient()
	require.NoError(t, err)

	_, err = c.BeginAuthSession(t.Context(), nil)
	require.Error(t, err)
}

func TestClient_SubmitSteamGuardCode_BeforeBegin(t *testing.T) {
	t.Parallel()

	c, err := client.NewClient()
	require.NoError(t, err)

	_, err = c.SubmitSteamGuardCode(t.Context(), &identity.AuthSession{}, "x")
	require.Error(t, err)
	require.Contains(t, err.Error(), "BeginAuthSession first")
}