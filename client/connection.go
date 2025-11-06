package client

import (
	"context"
	"fmt"
	"slices"

	"github.com/lewisgibson/go-steam/api/services/isteamdirectory"
	"github.com/lewisgibson/go-steam/connection"
	"github.com/lewisgibson/go-steam/protocol"
)

// Connect connects to the server.
func (c *Client) Connect(ctx context.Context) error {
	for {
		var err = c.connect(ctx)
		if !c.reconnect {
			return err
		}
		if err != nil {
			c.events <- &protocol.EventError{Err: err}
		}
	}
}

// connect sets up and runs the connection.
func (c *Client) connect(ctx context.Context) error {
	if _, err := c.identity.Identity(ctx); err != nil {
		return fmt.Errorf("failed to get identity: %w", err)
	}

	servers, err := c.api.SteamDirectory().GetCMListForConnect(ctx, isteamdirectory.GetCMListForConnectParameters{
		CMType: "netfilter", // TCP
		Realm:  "steamglobal",
	})
	switch {
	case err != nil:
		return fmt.Errorf("failed to get CM list for connect: %w", err)

	case len(servers) == 0:
		return fmt.Errorf("no servers found")
	}

	// Sort the servers by lowest load first
	slices.SortFunc(servers, func(a, b isteamdirectory.CMServer) int {
		return a.Load - b.Load
	})

	var conn protocol.Connection
	for _, server := range servers {
		switch server.Type {
		case "netfilter":
			conn, err = connection.NewTCPConnection(ctx, c.dialer, server.Endpoint)

		default:
			err = fmt.Errorf("unsupported server type: %s", server.Type)
		}
		if err != nil {
			continue
		}
	}
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	incoming := make(chan protocol.Event, 128)
	c.protocol, err = protocol.NewProtocol(conn, incoming)
	if err != nil {
		return fmt.Errorf("failed to create protocol: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-incoming:
				if !ok {
					return
				}

				select {
				case c.events <- event:
				default:
				}
			}
		}
	}()

	return c.protocol.Run(ctx)
}

// Disconnect disconnects from the server.
func (c *Client) Disconnect() error {
	if c.protocol == nil {
		return nil
	}
	c.reconnect = false
	if err := c.protocol.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
