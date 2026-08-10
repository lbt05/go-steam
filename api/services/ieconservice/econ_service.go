package ieconservice

import (
	"github.com/lbt05/go-steam/api/transports"
)

// IEconService is the interface for the IEconService API
type IEconService struct {
	// Transport is the transport for the IEconService service.
	//
	//go:generate mockgen -package=ieconservice_test -destination=mock_transport_test.go github.com/lbt05/go-steam/api/transports Transport
	transport transports.Transport
}

// NewIEconService creates a new IEconService
func NewIEconService(transport transports.Transport) *IEconService {
	// The transport is validated here in case this package is used directly, outside of the API wrapper which validates and returns an error.
	// Panicking is intentional since this runs during lazy service initialization, where propagating errors would decrease usability.
	if transport == nil {
		panic("transport is nil")
	}
	return &IEconService{transport: transport}
}
