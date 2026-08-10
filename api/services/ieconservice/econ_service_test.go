package ieconservice_test

import (
	"testing"

	"github.com/lbt05/go-steam/api/services/ieconservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIEconService_NilTransport(t *testing.T) {
	t.Parallel()

	// Assert: should panic if transport is nil
	require.Panics(t, func() {
		ieconservice.NewIEconService(nil)
	})
}

func TestIEconService_ValidTransport(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport
	transport := NewMockTransport(gomock.NewController(t))

	// Assert: should not panic if transport is provided
	require.NotPanics(t, func() {
		ieconservice.NewIEconService(transport)
	})
}
