package connection_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/connection"
	"github.com/stretchr/testify/assert"
)

// TestConnectionState verifies that connection state constants maintain their exact numeric values.
// This test prevents accidental changes to state values that could cause unexpected behavior.
//
// These numeric values should remain stable:
//   - StateDisconnecting = 0
//   - StateConnectedUnencrypted = 1
//   - StateConnectedEncrypted = 2
//
// Changing these values could break existing code that relies on these specific numeric values.
// If you need to add new states, append them after StateConnectedEncrypted.
func TestConnectionState(t *testing.T) {
	t.Parallel()

	// Assert: Verify exact numeric values to prevent accidental changes
	assert.Equal(t, connection.State(0), connection.StateDisconnecting)        //nolint:testifylint // Testing exact numeric values
	assert.Equal(t, connection.State(1), connection.StateConnectedUnencrypted) //nolint:testifylint // Testing exact numeric values
	assert.Equal(t, connection.State(2), connection.StateConnectedEncrypted)   //nolint:testifylint // Testing exact numeric values
}

// TestConnectionState_String verifies that the String method returns correct human-readable names.
// This ensures consistent string representations for logging and debugging purposes.
func TestConnectionState_String(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		state    connection.State
		expected string
	}{
		{connection.StateDisconnecting, "Disconnecting"},
		{connection.StateConnectedUnencrypted, "ConnectedUnencrypted"},
		{connection.StateConnectedEncrypted, "ConnectedEncrypted"},
		{connection.State(99), "Unknown"}, // Test unknown state
	}
	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()

			// Assert: Verify the string representation matches expected
			assert.Equal(t, tc.expected, tc.state.String())
		})
	}
}
