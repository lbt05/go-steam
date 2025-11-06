package connection

// State represents the current state of a Steam connection.
// The state indicates the connection's lifecycle phase and determines
// which operations are valid to perform on the connection.
//
// State transitions follow a specific lifecycle:
//   - New connections start in StateConnectedUnencrypted
//   - After session key exchange, they transition to StateConnectedEncrypted
//   - When closing, they transition to StateDisconnecting
//
// This type is thread-safe when accessed through connection methods.
type State uint8

const (
	// StateDisconnecting indicates the connection is being closed.
	// In this state:
	//   - No new operations should be started
	//   - Existing operations may return nil to indicate graceful shutdown
	//   - The underlying network connection is being closed
	StateDisconnecting State = iota

	// StateConnectedUnencrypted indicates an active unencrypted connection.
	// In this state:
	//   - Basic protocol operations are available
	//   - Data is transmitted in plaintext
	//   - Session key exchange can be performed to upgrade to encrypted state
	//   - This is the initial state for new connections
	StateConnectedUnencrypted

	// StateConnectedEncrypted indicates an active encrypted connection.
	// In this state:
	//   - All protocol operations are available
	//   - Data is encrypted using the session key
	//   - This is the preferred state for Steam protocol communication
	//   - Sensitive operations require this state
	StateConnectedEncrypted
)

// String returns a human-readable representation of the connection state.
// This is useful for debugging and logging purposes.
func (s State) String() string {
	switch s {
	case StateDisconnecting:
		return "Disconnecting"

	case StateConnectedUnencrypted:
		return "ConnectedUnencrypted"

	case StateConnectedEncrypted:
		return "ConnectedEncrypted"

	default:
		return "Unknown"
	}
}
