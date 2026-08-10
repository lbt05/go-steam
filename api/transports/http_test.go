package transports_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/lbt05/go-steam/api/transports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewHTTPTransport_DefaultValues(t *testing.T) {
	t.Parallel()

	// Act: create a new HTTP transport with no options
	transport := transports.NewHTTPTransport()

	// Assert: the transport should have default values
	require.NotNil(t, transport)
	assert.Equal(t, "https://api.steampowered.com", transport.BaseURL)
	assert.Equal(t, "go-steam", transport.UserAgent)
	assert.NotNil(t, transport.HttpClient)
	assert.Nil(t, transport.Tracer)
}

func TestNewHTTPTransport_WithBaseURL(t *testing.T) {
	t.Parallel()

	// Act: create a new HTTP transport with a custom base URL
	transport := transports.NewHTTPTransport(
		transports.WithBaseURL("https://custom.api.com"),
	)

	// Assert: the transport should use the custom base URL
	require.NotNil(t, transport)
	assert.Equal(t, "https://custom.api.com", transport.BaseURL)
}

func TestNewHTTPTransport_WithUserAgent(t *testing.T) {
	t.Parallel()

	// Act: create a new HTTP transport with a custom user agent
	transport := transports.NewHTTPTransport(
		transports.WithUserAgent("custom-agent"),
	)

	// Assert: the transport should use the custom user agent
	require.NotNil(t, transport)
	assert.Equal(t, "custom-agent", transport.UserAgent)
}

func TestNewHTTPTransport_WithHTTPClient(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Act: create a new HTTP transport with a custom HTTP client
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Assert: the transport should use the custom HTTP client
	require.NotNil(t, transport)
	assert.Equal(t, mockClient, transport.HttpClient)
}

func TestNewHTTPTransport_WithMultipleOptions(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Act: create a new HTTP transport with multiple options
	transport := transports.NewHTTPTransport(
		transports.WithBaseURL("https://custom.api.com"),
		transports.WithUserAgent("custom-agent"),
		transports.WithHTTPClient(mockClient),
	)

	// Assert: the transport should use all custom values
	require.NotNil(t, transport)
	assert.Equal(t, "https://custom.api.com", transport.BaseURL)
	assert.Equal(t, "custom-agent", transport.UserAgent)
	assert.Equal(t, mockClient, transport.HttpClient)
}

func TestHTTPTransport_Call_GetRequest_Success(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create test parameters
	type testParams struct {
		Key string `url:"key"`
	}
	params := testParams{Key: "value"}

	// Arrange: create expected response
	type testResponse struct {
		Result string `json:"result"`
	}
	expectedResponse := testResponse{Result: "success"}
	responseBody, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	// Arrange: setup mock expectations
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the request should be a GET request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Contains(t, req.URL.String(), "key=value")
		assert.Contains(t, req.URL.String(), "format=json")
		assert.Contains(t, req.URL.String(), "/TestService/TestMethod/v1/")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
			Header:     http.Header{},
		}, nil
	})

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response testResponse
	err = transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, params, &response)

	// Assert: the call should succeed
	require.NoError(t, err)
	assert.Equal(t, "success", response.Result)
}

func TestHTTPTransport_Call_PostRequest_Success(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create test parameters
	type testParams struct {
		Key string `url:"key"`
	}
	params := testParams{Key: "value"}

	// Arrange: create expected response
	type testResponse struct {
		Result string `json:"result"`
	}
	expectedResponse := testResponse{Result: "success"}
	responseBody, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	// Arrange: setup mock expectations
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the request should be a POST request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

		// Assert: the body should contain the parameters
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "key=value")
		assert.Contains(t, string(body), "format=json")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
			Header:     http.Header{},
		}, nil
	})

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response testResponse
	err = transport.Call(context.Background(), http.MethodPost, "TestService", "TestMethod", 1, params, &response)

	// Assert: the call should succeed
	require.NoError(t, err)
	assert.Equal(t, "success", response.Result)
}

func TestHTTPTransport_Call_NilParams(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create expected response
	type testResponse struct {
		Result string `json:"result"`
	}
	expectedResponse := testResponse{Result: "success"}
	responseBody, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	// Arrange: setup mock expectations
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the request should still contain format=json
		assert.Contains(t, req.URL.String(), "format=json")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
			Header:     http.Header{},
		}, nil
	})

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API with nil params
	var response testResponse
	err = transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should succeed
	require.NoError(t, err)
	assert.Equal(t, "success", response.Result)
}

func TestHTTPTransport_Call_InvalidParams(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create invalid params (channel cannot be encoded)
	invalidParams := make(chan int)

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API with invalid params
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, invalidParams, &response)

	// Assert: the call should fail with query encoding error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create query values")
}

func TestHTTPTransport_Call_HTTPClientError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: setup mock to return an error
	mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("network error"))

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should fail with network error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
	assert.Contains(t, err.Error(), "network error")
}

func TestHTTPTransport_Call_EResultHeaderError(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: setup mock to return a response with error header
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header: http.Header{
			"X-Eresult": []string{"42"},
		},
	}, nil)

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should fail with error code
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed with error code: 42")
}

func TestHTTPTransport_Call_EResultHeaderSuccess(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create expected response
	type testResponse struct {
		Result string `json:"result"`
	}
	expectedResponse := testResponse{Result: "success"}
	responseBody, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	// Arrange: setup mock to return a response with success header (1)
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
		Header: http.Header{
			"X-Eresult": []string{"1"},
		},
	}, nil)

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response testResponse
	err = transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should succeed even with X-Eresult=1
	require.NoError(t, err)
	assert.Equal(t, "success", response.Result)
}

func TestHTTPTransport_Call_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: setup mock to return invalid JSON
	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
		Header:     http.Header{},
	}, nil)

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should fail with JSON decode error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestHTTPTransport_Call_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Arrange: setup mock to return context cancelled error
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the request context should be cancelled
		assert.Error(t, req.Context().Err()) //nolint:testifylint // Assert inside callbacks.
		return nil, req.Context().Err()
	})

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API with cancelled context
	var response map[string]any
	err := transport.Call(ctx, http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should fail with context cancellation error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to make request")
}

func TestHTTPTransport_Call_URLConstruction(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name            string
		verb            string
		service         string
		method          string
		version         int
		expectedURLPart string
	}{
		{
			name:            "GET with version 1",
			verb:            http.MethodGet,
			service:         "ISteamDirectory",
			method:          "GetCMList",
			version:         1,
			expectedURLPart: "/ISteamDirectory/GetCMList/v1/",
		},
		{
			name:            "POST with version 2",
			verb:            http.MethodPost,
			service:         "IAuthenticationService",
			method:          "BeginAuthSession",
			version:         2,
			expectedURLPart: "/IAuthenticationService/BeginAuthSession/v2",
		},
		{
			name:            "GET with version 100",
			verb:            http.MethodGet,
			service:         "TestService",
			method:          "TestMethod",
			version:         100,
			expectedURLPart: "/TestService/TestMethod/v100/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: create a mock controller and HTTP client
			ctrl := gomock.NewController(t)
			mockClient := NewMockHTTPClient(ctrl)

			// Arrange: setup mock expectations
			mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
				// Assert: the URL should contain the expected part
				assert.Contains(t, req.URL.String(), tc.expectedURLPart)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{},
				}, nil
			})

			// Arrange: create transport
			transport := transports.NewHTTPTransport(
				transports.WithHTTPClient(mockClient),
			)

			// Act: call the API
			var response map[string]any
			err := transport.Call(context.Background(), tc.verb, tc.service, tc.method, tc.version, nil, &response)

			// Assert: the call should succeed
			require.NoError(t, err)
		})
	}
}

func TestHTTPTransport_Call_CustomBaseURL(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	customBaseURL := "https://custom.example.com"

	// Arrange: setup mock expectations
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the URL should start with the custom base URL
		assert.True(t, strings.HasPrefix(req.URL.String(), customBaseURL))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     http.Header{},
		}, nil
	})

	// Arrange: create transport with custom base URL
	transport := transports.NewHTTPTransport(
		transports.WithBaseURL(customBaseURL),
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, nil, &response)

	// Assert: the call should succeed
	require.NoError(t, err)
}

func TestHTTPTransport_Call_SpecialCharactersInParams(t *testing.T) {
	t.Parallel()

	// Arrange: create a mock controller and HTTP client
	ctrl := gomock.NewController(t)
	mockClient := NewMockHTTPClient(ctrl)

	// Arrange: create params with special characters
	type testParams struct {
		Key string `url:"key"`
	}
	params := testParams{Key: "value with spaces & special=chars"}

	// Arrange: setup mock expectations
	mockClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
		// Assert: the special characters should be properly URL encoded
		if req.Method == http.MethodGet {
			parsedURL, err := url.Parse(req.URL.String())
			require.NoError(t, err)
			assert.Equal(t, "value with spaces & special=chars", parsedURL.Query().Get("key"))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     http.Header{},
		}, nil
	})

	// Arrange: create transport
	transport := transports.NewHTTPTransport(
		transports.WithHTTPClient(mockClient),
	)

	// Act: call the API
	var response map[string]any
	err := transport.Call(context.Background(), http.MethodGet, "TestService", "TestMethod", 1, params, &response)

	// Assert: the call should succeed
	require.NoError(t, err)
}
