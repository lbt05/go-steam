package iplayerservice

import (
	"github.com/lbt05/go-steam/api/transports"
)

// IPlayerService is the interface for the IPlayerService API.
type IPlayerService struct {
	// Transport is the transport for the IPlayerService service.
	//
	//go:generate mockgen -package=iplayerservice_test -destination=mock_transport_test.go github.com/lbt05/go-steam/api/transports Transport
	transport transports.Transport
}

// NewIPlayerService creates a new IPlayerService.
func NewIPlayerService(transport transports.Transport) *IPlayerService {
	// The transport is validated here in case this package is used directly, outside of the API wrapper which validates and returns an error.
	// Panicking is intentional since this runs during lazy service initialization, where propagating errors would decrease usability.
	if transport == nil {
		panic("transport is nil")
	}
	return &IPlayerService{transport: transport}
}