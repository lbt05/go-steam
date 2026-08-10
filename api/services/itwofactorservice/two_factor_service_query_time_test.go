package itwofactorservice_test

import (
	"errors"
	"testing"

	"github.com/lbt05/go-steam/api/services/itwofactorservice"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestITwoFactorService_QueryTime_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns an error
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("network error"))

	// Arrange: create service
	service := itwofactorservice.NewITwoFactorService(transport)

	// Act: query time
	response, err := service.QueryTime(t.Context())

	// Assert: should fail with error
	require.Error(t, err)
	require.Nil(t, response)
}
