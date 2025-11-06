package transports

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HTTPClient is a client that can perform HTTP requests.
//
//go:generate mockgen -package=transports_test -destination=mock_http_client_test.go github.com/lewisgibson/go-steam/api/transports HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPTransport is a transport that can be used to call the API.
type HTTPTransport struct {
	// BaseURL is the base URL for the API.
	BaseURL string
	// UserAgent is the user agent for the API.
	UserAgent string
	// HttpClient is the HTTP client for the API.
	HttpClient HTTPClient
	// Tracer is the tracer for the HTTP transport.
	Tracer trace.Tracer
}

// Compile time check that HTTPTransport implements the Transport interface.
var _ Transport = (*HTTPTransport)(nil)

// Option is a function that can be used to configure the HTTP transport.
type Option func(*HTTPTransport)

// WithBaseURL sets the base URL for the HTTP transport.
func WithBaseURL(baseURL string) Option {
	return func(o *HTTPTransport) {
		o.BaseURL = baseURL
	}
}

// WithUserAgent sets the user agent for the HTTP transport.
func WithUserAgent(userAgent string) Option {
	return func(o *HTTPTransport) {
		o.UserAgent = userAgent
	}
}

// WithHTTPClient sets the HTTP client for the HTTP transport.
func WithHTTPClient(client HTTPClient) Option {
	return func(o *HTTPTransport) {
		o.HttpClient = client
	}
}

// NewHTTPTransport creates a new HTTP transport.
func NewHTTPTransport(options ...Option) *HTTPTransport {
	var transport = &HTTPTransport{
		BaseURL:    "https://api.steampowered.com",
		UserAgent:  "go-steam",
		HttpClient: http.DefaultClient,
	}
	for _, option := range options {
		option(transport)
	}
	return transport
}

// Call calls the API with the given verb, service, method, version, and params.
func (t *HTTPTransport) Call(ctx context.Context, verb string, service, method string, version int, params any, response any) error {
	if t.Tracer != nil {
		var span trace.Span
		ctx, span = t.Tracer.Start(ctx, fmt.Sprintf("%s %s %s v%d", verb, service, method, version),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("base_url", t.BaseURL),
				attribute.String("user_agent", t.UserAgent),

				attribute.String("verb", verb),
				attribute.String("service", service),
				attribute.String("method", method),
				attribute.Int("version", version),
			),
		)
		defer span.End()
	}

	// The values are encoded as url parameters for get requests.
	var values = url.Values{}
	if params != nil {
		var err error
		values, err = query.Values(params)
		if err != nil {
			return fmt.Errorf("failed to create query values: %w", err)
		}
	}
	values.Add("format", "json")

	// The values are encoded as url parameters for get requests.
	var url = fmt.Sprintf("%s/%s/%s/v%d", t.BaseURL, service, method, version)
	if verb == http.MethodGet {
		url = fmt.Sprintf("%s/%s/%s/v%d/?%s", t.BaseURL, service, method, version, values.Encode())
	}

	// The values are encoded as a body for post requests.
	var body io.Reader = http.NoBody
	if verb == http.MethodPost {
		body = strings.NewReader(values.Encode())
	}

	// The request is created.
	req, err := http.NewRequestWithContext(ctx, verb, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// The content type is set to application/x-www-form-urlencoded for post requests.
	if verb == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// if b, err := httputil.DumpRequest(req, true); err == nil {
	// 	log.Println(string(b))
	// }

	// The request is made.
	res, err := t.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	// if b, err := httputil.DumpResponse(res, true); err == nil {
	// 	log.Println(string(b))
	// }

	defer res.Body.Close()

	// Check for error in headers
	if eresult := res.Header.Get("X-Eresult"); eresult != "" && eresult != "1" {
		return fmt.Errorf("request failed with error code: %s", eresult)
	}

	// The response is decoded into the response pointer.
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
