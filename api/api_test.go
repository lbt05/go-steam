package api_test

import (
	"testing"

	"github.com/lewisgibson/go-steam/api"
	"github.com/lewisgibson/go-steam/api/transports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -package=api_test -destination=mock_transport_test.go github.com/lewisgibson/go-steam/api/transports Transport

func TestNewAPI_NilTransport(t *testing.T) {
	t.Parallel()

	// Act: create a new API with nil transport
	apiClient, err := api.NewAPI(nil)

	// Assert: should return an error
	require.Error(t, err)
	require.ErrorIs(t, err, api.ErrNilTransport)
	require.Nil(t, apiClient)
}

func TestNewAPI_ValidTransport(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport
	transport := NewMockTransport(gomock.NewController(t))

	// Act: create a new API with valid transport
	apiClient, err := api.NewAPI(transport)

	// Assert: should succeed
	require.NoError(t, err)
	require.NotNil(t, apiClient)
}

func TestAPI_SteamDirectory(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport and API client
	transport := NewMockTransport(gomock.NewController(t))
	apiClient, err := api.NewAPI(transport)
	require.NoError(t, err)

	// Act: get the SteamDirectory service twice
	steamDirectory1 := apiClient.SteamDirectory()
	steamDirectory2 := apiClient.SteamDirectory()

	// Assert: should return valid and consistent instance (lazy singleton)
	require.NotNil(t, steamDirectory1)
	require.NotNil(t, steamDirectory2)
	assert.Equal(t, steamDirectory1, steamDirectory2)
}

func TestAPI_IAuthenticationService(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport and API client
	transport := NewMockTransport(gomock.NewController(t))
	apiClient, err := api.NewAPI(transport)
	require.NoError(t, err)

	// Act: get the IAuthenticationService service twice
	authService1 := apiClient.IAuthenticationService()
	authService2 := apiClient.IAuthenticationService()

	// Assert: should return valid and consistent instance (lazy singleton)
	require.NotNil(t, authService1)
	require.NotNil(t, authService2)
	assert.Equal(t, authService1, authService2)
}

func TestAPI_ITwoFactorService(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock transport and API client
	transport := NewMockTransport(gomock.NewController(t))
	apiClient, err := api.NewAPI(transport)
	require.NoError(t, err)

	// Act: get the ITwoFactorService service twice
	twoFactorService1 := apiClient.ITwoFactorService()
	twoFactorService2 := apiClient.ITwoFactorService()

	// Assert: should return valid and consistent instance (lazy singleton)
	require.NotNil(t, twoFactorService1)
	require.NotNil(t, twoFactorService2)
	assert.Equal(t, twoFactorService1, twoFactorService2)
}

func TestAPI_WithHTTPTransport(t *testing.T) {
	t.Parallel()

	// Arrange: create an HTTP transport
	httpTransport := transports.NewHTTPTransport()

	// Act: create API client with HTTP transport
	apiClient, err := api.NewAPI(httpTransport)
	require.NoError(t, err)

	// Assert: all services should be accessible
	require.NotNil(t, apiClient.SteamDirectory())
	require.NotNil(t, apiClient.IAuthenticationService())
	require.NotNil(t, apiClient.ITwoFactorService())
}
