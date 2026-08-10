package iauthenticationservice_test

import (
	"errors"
	"testing"

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIAuthenticationService_UpdateAuthSessionWithSteamGuardCode_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns an error
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("transport error"))

	// Arrange: create service
	service := iauthenticationservice.NewIAuthenticationService(transport)

	// Act: update auth session with steam guard code
	response, err := service.UpdateAuthSessionWithSteamGuardCode(t.Context(), iauthenticationservice.UpdateAuthSessionWithSteamGuardCodeParameters{})

	// Assert: should fail with error
	require.Error(t, err)
	require.Nil(t, response)
}
