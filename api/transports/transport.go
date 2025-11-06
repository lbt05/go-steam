package transports

import (
	"context"
)

// Transport is a transport that can be used to call the API.
//
//go:generate mockgen -package=transports_test -destination=mock_transport_test.go github.com/lewisgibson/go-steam/api/transports Transport
type Transport interface {
	// Call calls the API with the given verb, service, method, version, and params.
	Call(ctx context.Context, verb string, service, method string, version int, params any, response any) error
}
