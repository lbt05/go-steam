package itwofactorservice

import (
	"github.com/lewisgibson/go-steam/api/transports"
)

// ITwoFactorService is the interface for the ITwoFactorService API
type ITwoFactorService struct {
	// Transport is the transport for the ITwoFactorService service.
	//
	//go:generate mockgen -package=itwofactorservice_test -destination=mock_transport_test.go github.com/lewisgibson/go-steam/api/transports Transport
	transport transports.Transport
}

// NewITwoFactorService creates a new ITwoFactorService
func NewITwoFactorService(transport transports.Transport) *ITwoFactorService {
	// The transport is validated here in case this package is used directly, outside of the API wrapper which validates and returns an error.
	// Panicking is intentional since this runs during lazy service initialization, where propagating errors would decrease usability.
	if transport == nil {
		panic("transport is nil")
	}
	return &ITwoFactorService{transport: transport}
}
