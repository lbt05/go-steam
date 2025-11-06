package connection

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/lewisgibson/go-steam/crypto"
)

// Common errors for TCP connection operations.
var (
	ErrNilContext        = errors.New("context cannot be nil")
	ErrNilDialer         = errors.New("dialer cannot be nil")
	ErrEmptyAddress      = errors.New("address cannot be empty")
	ErrEmptyData         = errors.New("data cannot be empty")
	ErrConnectionClosing = errors.New("connection is disconnecting")
	ErrInvalidState      = errors.New("invalid connection state")
	ErrConnectionClosed  = errors.New("connection closed by remote peer")
	ErrPacketTooLarge    = errors.New("packet too large")
	ErrInvalidMagic      = errors.New("invalid magic number")
	ErrIncompletePacket  = errors.New("incomplete packet")
)

// TCPMagic is the magic number used in Steam TCP protocol packets.
// It serves as a protocol identifier to ensure packet synchronization.
const TCPMagic uint32 = 0x31305456

// TCPConnection represents a TCP connection to a Steam server.
// It handles the Steam-specific TCP protocol including packet framing,
// encryption/decryption, and state management.
type TCPConnection struct {
	// mutex protects concurrent access to state and key fields
	mutex sync.RWMutex
	// state represents the current connection state (disconnecting, unencrypted, encrypted)
	state State
	// connection is the underlying network connection
	connection net.Conn
	// sessionKey is the session sessionKey used for encryption/decryption
	sessionKey []byte
}

// NewTCPConnection creates a new TCP connection to the specified address.
// The connection starts in StateConnectedUnencrypted state and must be
// upgraded to encrypted state using SetSessionKey before sending sensitive data.
//
// The context is used for the initial connection establishment and can be
// cancelled to abort the connection attempt.
func NewTCPConnection(ctx context.Context, dialer Dialer, addr string) (*TCPConnection, error) {
	switch {
	case ctx == nil:
		return nil, ErrNilContext

	case dialer == nil:
		return nil, ErrNilDialer

	case addr == "":
		return nil, ErrEmptyAddress
	}

	connection, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return &TCPConnection{
		state:      StateConnectedUnencrypted,
		connection: connection,
	}, nil
}

// GetState returns the current connection state.
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) GetState() State {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.state
}

// SetSessionKey sets the session key for the connection and transitions
// the state to StateConnectedEncrypted. The session key is used for
// encrypting/decrypting all subsequent data sent through this connection.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) SetSessionKey(sessionKey []byte) {
	if len(sessionKey) == 0 {
		return // Don't set empty session key
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Only transition to encrypted state if we're currently unencrypted
	if c.state == StateConnectedUnencrypted {
		c.state = StateConnectedEncrypted
	}

	// Create a copy of the session key to avoid external modifications
	c.sessionKey = append(make([]byte, 0, len(sessionKey)), sessionKey...)
}

// GetSessionKey returns the session key for the connection.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) GetSessionKey() []byte {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return append(make([]byte, 0, len(c.sessionKey)), c.sessionKey...)
}

// Send sends data through the TCP connection.
// If a session key is set, the data will be encrypted before transmission.
// The data is sent using the Steam TCP protocol format: [length][magic][data]
// The context can be used to cancel the operation or set timeouts.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) Send(ctx context.Context, data []byte) error {
	switch {
	case ctx == nil:
		return ErrNilContext

	case len(data) == 0:
		return ErrEmptyData

	case c.GetState() == StateDisconnecting:
		return ErrConnectionClosing
	}

	// Encrypt data if session key is available
	if sessionKey := c.GetSessionKey(); len(sessionKey) != 0 {
		encrypted, err := crypto.SymmetricEncryptWithHmacIv(data, sessionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		data = encrypted
	}

	// Set write deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.connection.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
	}

	// Write packet length
	if err := binary.Write(c.connection, binary.LittleEndian, uint32(len(data))); err != nil {
		if errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting {
			return nil // Graceful shutdown
		}
		return fmt.Errorf("failed to write packet length: %w", err)
	}

	// Write magic number
	if err := binary.Write(c.connection, binary.LittleEndian, TCPMagic); err != nil {
		if errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting {
			return nil // Graceful shutdown
		}
		return fmt.Errorf("failed to write magic number: %w", err)
	}

	// Write data
	if _, err := c.connection.Write(data); err != nil {
		if errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting {
			return nil // Graceful shutdown
		}
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// Read reads a complete packet from the TCP connection.
// The method blocks until a complete packet is received or the context is cancelled.
// If a session key is set, the received data will be decrypted automatically.
//
// Returns a bytes.Reader containing the packet data, or nil if the connection
// is being gracefully closed. The context can be used to cancel the operation
// or set timeouts.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) Read(ctx context.Context) (*bytes.Reader, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}

	// Check if connection is in a valid state for reading
	if state := c.GetState(); state == StateDisconnecting {
		return nil, nil
	}

	// Set read deadline based on context
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.connection.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	// Read packet length
	var packetLength uint32
	switch err := binary.Read(c.connection, binary.LittleEndian, &packetLength); {
	case errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting:
		return nil, nil

	case errors.Is(err, io.EOF):
		return nil, fmt.Errorf("%w: %w", ErrConnectionClosed, err)

	case err != nil:
		return nil, fmt.Errorf("failed to read packet length: %w", err)

	case packetLength > 1024*1024:
		return nil, fmt.Errorf("%w: %d bytes (max: %d)", ErrPacketTooLarge, packetLength, 1024*1024)
	}

	// Read magic
	var packetMagic uint32
	switch err := binary.Read(c.connection, binary.LittleEndian, &packetMagic); {
	case errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting:
		return nil, nil

	case errors.Is(err, io.EOF):
		return nil, fmt.Errorf("%w: %w", ErrConnectionClosed, err)

	case err != nil:
		return nil, fmt.Errorf("failed to read magic: %w", err)

	case packetMagic != TCPMagic:
		return nil, fmt.Errorf("%w: got 0x%x, expected 0x%x (connection out of sync)", ErrInvalidMagic, packetMagic, TCPMagic)
	}

	// Read packet body
	var bodyBytes = make([]byte, packetLength)
	switch n, err := io.ReadFull(c.connection, bodyBytes); {
	case errors.Is(err, net.ErrClosed) && c.GetState() == StateDisconnecting:
		return nil, nil

	case errors.Is(err, io.EOF):
		return nil, fmt.Errorf("%w: %w", ErrConnectionClosed, err)

	case err != nil:
		return nil, fmt.Errorf("failed to read packet body: %w", err)

	case n < int(packetLength):
		return nil, fmt.Errorf("%w: read %d bytes, expected %d", ErrIncompletePacket, n, packetLength)
	}

	// Decrypt data if session key is available
	if sessionKey := c.GetSessionKey(); len(sessionKey) != 0 {
		decrypted, err := crypto.SymmetricDecrypt(bodyBytes, sessionKey, true)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt packet: %w", err)
		}
		bodyBytes = decrypted
	}

	return bytes.NewReader(bodyBytes), nil
}

// Close gracefully closes the TCP connection.
// The connection state is set to StateDisconnecting and the underlying
// network connection is closed. This method is idempotent and can be
// called multiple times safely.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) Close() error {
	c.mutex.Lock()
	// Only transition to disconnecting if not already disconnecting
	if c.state != StateDisconnecting {
		c.state = StateDisconnecting
	}
	c.mutex.Unlock()

	// Close the underlying connection
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// IsConnected returns true if the connection is in a connected state
// (either encrypted or unencrypted), false otherwise.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) IsConnected() bool {
	var state = c.GetState()
	return state == StateConnectedUnencrypted || state == StateConnectedEncrypted
}

// IsEncrypted returns true if the connection is in encrypted state, false otherwise.
//
// This method is thread-safe and can be called concurrently.
func (c *TCPConnection) IsEncrypted() bool {
	return c.GetState() == StateConnectedEncrypted
}

// RemoteAddr returns the remote network address of the connection.
// Returns nil if the connection is not established.
func (c *TCPConnection) RemoteAddr() net.Addr {
	return c.connection.RemoteAddr()
}

// LocalAddr returns the local network address of the connection.
// Returns nil if the connection is not established.
func (c *TCPConnection) LocalAddr() net.Addr {
	return c.connection.LocalAddr()
}
