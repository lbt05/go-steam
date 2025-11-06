package iauthenticationservice_test

import (
	"errors"
	"testing"

	"github.com/lewisgibson/go-steam/api/services/iauthenticationservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIAuthenticationService_BeginAuthSessionViaCredentials_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns an error
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("transport error"))

	// Arrange: create service
	service := iauthenticationservice.NewIAuthenticationService(transport)

	// Act: begin auth session
	response, err := service.BeginAuthSessionViaCredentials(t.Context(), iauthenticationservice.BeginAuthSessionViaCredentialsParameters{})

	// Assert: should fail with error
	require.Error(t, err)
	require.Nil(t, response)
}
