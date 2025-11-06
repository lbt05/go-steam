package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/lewisgibson/go-steam/api"
	"github.com/lewisgibson/go-steam/api/services/isteamdirectory"
	"github.com/lewisgibson/go-steam/api/transports"
	"github.com/lewisgibson/go-steam/connection"
	"github.com/lewisgibson/go-steam/protocol"
	"github.com/lewisgibson/go-steam/steamid"
)

// Identity is the interface for identities.
//
//go:generate mockgen -destination=mock_identity_test.go -package=client_test . Identity
type Identity interface {
	SteamID() steamid.SteamID
	RefreshToken() string
}

// IdentityProvider is the interface for identity providers.
//
//go:generate mockgen -destination=mock_identity_provider_test.go -package=client_test . IdentityProvider
type IdentityProvider interface {
	Identity(ctx context.Context) (Identity, error)
}

type API interface {
	SteamDirectory() *isteamdirectory.ISteamDirectory
}

// Client represents a Steam client.
type Client struct {
	api      API
	dialer   connection.Dialer
	events   chan protocol.Event
	identity IdentityProvider
	protocol *protocol.Protocol

	reconnect bool
}

// ClientOption is a function that can be used to configure the client.
type ClientOption func(*Client)

// WithAPI sets the API for the client.
func WithAPI(api API) ClientOption {
	return func(c *Client) {
		c.api = api
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

// NewClient creates a new Steam client.
func NewClient(identity IdentityProvider, opts ...ClientOption) (*Client, error) {
	var client = &Client{
		identity: identity,
	}
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
