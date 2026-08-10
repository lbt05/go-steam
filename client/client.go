package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lbt05/go-steam/api"
	"github.com/lbt05/go-steam/api/services/isteamdirectory"
	"github.com/lbt05/go-steam/api/transports"
	"github.com/lbt05/go-steam/connection"
	"github.com/lbt05/go-steam/identity"
	"github.com/lbt05/go-steam/protocol"
	"github.com/lbt05/go-steam/steamid"
)

// Identity is the interface for identities returned after a successful
// Steam Guard code submission.
//
//go:generate mockgen -destination=mock_identity_test.go -package=client_test . Identity
type Identity interface {
	SteamID() steamid.SteamID
	RefreshToken() string
}

type API interface {
	SteamDirectory() *isteamdirectory.ISteamDirectory
}

// Client represents a Steam client.
type Client struct {
	api       API
	dialer    connection.Dialer
	events    chan protocol.Event
	protocol  *protocol.Protocol
	reconnect bool

	authMu       sync.Mutex
	authIDP      *identity.CredentialsIdentityProvider
	authIdentity *identity.Identity

	// authAPI is an optional override for the identity API used during
	// two-step authentication. If unset, a default HTTP-backed API is
	// created lazily by BeginAuthSession.
	authAPI identity.API
}

// ClientOption is a function that can be used to configure the client.
type ClientOption func(*Client)

// WithAPI sets the API for the client.
func WithAPI(api API) ClientOption {
	return func(c *Client) {
		c.api = api
	}
}

// WithAuthAPI sets the identity API used for two-step authentication.
// When unset, BeginAuthSession builds a default identity.API backed by an
// HTTP transport.
func WithAuthAPI(api identity.API) ClientOption {
	return func(c *Client) {
		c.authAPI = api
	}
}

// WithDialer sets the dialer for the client.
func WithDialer(dialer connection.Dialer) ClientOption {
	return func(c *Client) {
		c.dialer = dialer
	}
}

// WithReconnect sets the reconnect flag for the client.
func WithReconnect(reconnect bool) ClientOption {
	return func(c *Client) {
		c.reconnect = reconnect
	}
}

// NewClient creates a new Steam client. Two-step authentication is initiated
// by calling Client.BeginAuthSession followed by Client.SubmitSteamGuardCode
// before Client.Connect.
func NewClient(opts ...ClientOption) (*Client, error) {
	var client = &Client{}
	for _, opt := range opts {
		opt(client)
	}
	if client.api == nil {
		api, err := api.NewAPI(transports.NewHTTPTransport())
		if err != nil {
			return nil, fmt.Errorf("failed to create API: %w", err)
		}
		client.api = api
	}
	if client.dialer == nil {
		client.dialer = &net.Dialer{
			Timeout: 10 * time.Second,
		}
	}
	if client.events == nil {
		client.events = make(chan protocol.Event, 512)
	}
	return client, nil
}

// BeginAuthSession performs the first half of the login flow: it encrypts
// the password with Steam's RSA public key and opens an authentication
// session. On success it also emits an EventSteamGuardChallenge on the
// client events channel so event-driven callers can react.
func (c *Client) BeginAuthSession(ctx context.Context, credentials *Credentials) (*identity.AuthSession, error) {
	if credentials == nil {
		return nil, errors.New("credentials are required")
	}

	c.authMu.Lock()
	defer c.authMu.Unlock()

	if c.authIDP == nil {
		var opts []identity.CredentialsIdentityProviderOption
		if c.authAPI != nil {
			opts = append(opts, identity.WithAPI(c.authAPI))
		}
		idp, err := identity.NewCredentialsIdentityProvider(identity.Credentials{
			AccountName:  credentials.Username,
			Password:     credentials.Password,
			SharedSecret: credentials.SharedSecret,
		}, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create identity provider: %w", err)
		}
		c.authIDP = idp
	}

	session, err := c.authIDP.BeginAuthSession(ctx)
	if err != nil {
		return nil, err
	}

	select {
	case c.events <- &protocol.EventSteamGuardChallenge{Session: session, AllowedConfirmations: session.AllowedConfirmations}:
	default:
	}

	return session, nil
}

// SubmitSteamGuardCode performs the second half of the login flow: it
// submits the user-supplied Steam Guard code to Steam, polls until a
// refresh token is issued, and caches the resulting identity on the
// client. The cached identity is returned by subsequent Identity calls
// (used by Client.Connect and Client.Logon).
//
// The code type (DeviceCode for TOTP, EmailCode for email 2FA, etc.) is
// chosen automatically from session.AllowedConfirmations.
func (c *Client) SubmitSteamGuardCode(ctx context.Context, session *identity.AuthSession, code string) (Identity, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	if c.authIDP == nil {
		return nil, errors.New("identity provider not initialized; call BeginAuthSession first")
	}

	id, err := c.authIDP.SubmitSteamGuardCode(ctx, session, code)
	if err != nil {
		return nil, err
	}
	c.authIdentity = id
	return id, nil
}

// Identity returns the identity cached by the most recent successful
// SubmitSteamGuardCode call. It is used by Connect and Logon to fetch
// the refresh token; calling it before authentication completes returns
// an error.
func (c *Client) Identity(_ context.Context) (Identity, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.authIdentity == nil {
		return nil, errors.New("identity not set; call BeginAuthSession and SubmitSteamGuardCode first")
	}
	return c.authIdentity, nil
}