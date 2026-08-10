package iauthenticationservice_test

import (
	"errors"
	"testing"

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIAuthenticationService_GetPasswordRSAPublicKey_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns an error
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("transport error"))

	// Arrange: create service
	service := iauthenticationservice.NewIAuthenticationService(transport)

	// Act: get password RSA public key
	response, err := service.GetPasswordRSAPublicKey(t.Context(), iauthenticationservice.GetPasswordRSAPublicKeyParameters{})

	// Assert: should fail with error
	require.Error(t, err)
	require.Nil(t, response)
}
