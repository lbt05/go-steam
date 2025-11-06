package protocol

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/lewisgibson/go-steam/language/steam"
	"github.com/lewisgibson/go-steam/steamid"
)

const (
	// JOBID_NONE is the none job id.
	JOBID_NONE = 18446744073709551615
	// PROTO_MASK is the mask for the protocol.
	PROTO_MASK uint32 = 0x80000000
)

// Header is the header for the protocol.
type Header struct {
	EMsg        steam.EMsg                `json:"EMsg,omitempty"`
	Proto       *steam.CMsgProtoBufHeader `json:"Proto,omitempty"`
	TargetJobID uint64                    `json:"TargetJobID,omitempty"`
	SourceJobID uint64                    `json:"SourceJobID,omitempty"`
	SteamID     uint64                    `json:"SteamID,omitempty"`
	SessionID   int32                     `json:"SessionID,omitempty"`
}

// ErrEmptyConnection is returned when a connection is empty.
var ErrEmptyConnection = errors.New("empty connection")

// Connection is the interface for the connection to the server.
//
//go:generate mockgen -destination=mock_connection_test.go -package=protocol . Connection
type Connection interface {
	Close() error
	SetSessionKey([]byte)
	Send(context.Context, []byte) error
	Read(context.Context) (*bytes.Reader, error)
}

// Protocol is the handler for the protocol.
type Protocol struct {
	// mutex protects concurrent access to events, connection, sessionKey, and stopHeartbeat
	mutex sync.Mutex
	// events is the channel for the events
	events chan<- Event
	// connection is the connection to the server
	connection Connection
	// sessionKey is the session key for the connection
	sessionKey []byte
	// stopHeartbeat is the channel for the stop heartbeat
	stopHeartbeat chan struct{}

	steamID   steamid.SteamID
	sessionID int32
}

// NewProtocol creates a new handler for the protocol.
func NewProtocol(connection Connection, events chan<- Event) (*Protocol, error) {
	if connection == nil {
		return nil, ErrEmptyConnection
	}
	return &Protocol{
		events:     events,
		connection: connection,
	}, nil
}

// Run runs the protocol.
func (h *Protocol) Run(ctx context.Context) error {
	for {
		// Read the packet from the connection.
		reader, err := h.connection.Read(ctx)
		switch {
		case err != nil:
			return fmt.Errorf("failed to read packet: %w", err)

		case reader == nil:
			return nil
		}

		// Read the message.
		if err := h.readMessage(ctx, reader); err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}
	}
}

// Close closes the protocol.
func (h *Protocol) Close() error {
	h.mutex.Lock()
	if h.stopHeartbeat != nil {
		close(h.stopHeartbeat)
		h.stopHeartbeat = nil
	}
	h.mutex.Unlock()
	return h.connection.Close()
}
