package iauthenticationservice

import (
	"github.com/lewisgibson/go-steam/api/transports"
)

// IAuthenticationService is the interface for the IAuthenticationService API
type IAuthenticationService struct {
	// Transport is the transport for the IAuthenticationService service.
	//
	//go:generate mockgen -package=iauthenticationservice_test -destination=mock_transport_test.go github.com/lewisgibson/go-steam/api/transports Transport
	transport transports.Transport
}

// NewIAuthenticationService creates a new IAuthenticationService
func NewIAuthenticationService(transport transports.Transport) *IAuthenticationService {
	// The transport is validated here in case this package is used directly, outside of the API wrapper which validates and returns an error.
	// Panicking is intentional since this runs during lazy service initialization, where propagating errors would decrease usability.
	if transport == nil {
		panic("transport is nil")
	}
	return &IAuthenticationService{transport: transport}
}
