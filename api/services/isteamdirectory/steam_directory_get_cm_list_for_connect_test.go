package isteamdirectory_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lewisgibson/go-steam/api/services/isteamdirectory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestISteamDirectory_GetCMListForConnect_Success(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), http.MethodGet, "ISteamDirectory", "GetCMListForConnect", 1, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, verb, service, method string, version int, params, response any) error {
			res := response.(*isteamdirectory.GetCMListForConnectResponse)
			res.Response.Success = true
			res.Response.ServerList = []isteamdirectory.CMServer{
				{Endpoint: "162.254.197.40:27017"},
			}
			return nil
		})

	// Arrange: create service
	service := isteamdirectory.NewISteamDirectory(transport)

	// Act: get CM list for connect
	servers, err := service.GetCMListForConnect(t.Context(), isteamdirectory.GetCMListForConnectParameters{})

	// Assert: should succeed and return servers
	require.NoError(t, err)
	require.Len(t, servers, 1)
	assert.Equal(t, "162.254.197.40:27017", servers[0].Endpoint)
}

func TestISteamDirectory_GetCMListForConnect_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns an error
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("network error"))

	// Arrange: create service
	service := isteamdirectory.NewISteamDirectory(transport)

	// Act: get CM list for connect
	servers, err := service.GetCMListForConnect(t.Context(), isteamdirectory.GetCMListForConnectParameters{})

	// Assert: should fail with error
	require.Error(t, err)
	require.Nil(t, servers)
}

func TestISteamDirectory_GetCMListForConnect_FailureResponse(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport that returns a failure response
	transport := NewMockTransport(gomock.NewController(t))
	transport.EXPECT().
		Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, verb, service, method string, version int, params, response any) error {
			res := response.(*isteamdirectory.GetCMListForConnectResponse)
			res.Response.Success = false
			res.Response.Message = "invalid request"
			return nil
		})

	// Arrange: create service
	service := isteamdirectory.NewISteamDirectory(transport)

	// Act: get CM list for connect
	servers, err := service.GetCMListForConnect(t.Context(), isteamdirectory.GetCMListForConnectParameters{})

	// Assert: should fail with error message
	require.Error(t, err)
	require.Nil(t, servers)
	assert.Contains(t, err.Error(), "invalid request")
}
