package isteamdirectory

import (
	"github.com/lbt05/go-steam/api/transports"
)

// ISteamDirectory is the ISteamDirectory service.
type ISteamDirectory struct {
	// Transport is the transport for the ISteamDirectory service.
	//
	//go:generate mockgen -package=isteamdirectory_test -destination=mock_transport_test.go github.com/lbt05/go-steam/api/transports Transport
	transport transports.Transport
}

// NewISteamDirectory creates a new ISteamDirectory instance with the given transport.
func NewISteamDirectory(transport transports.Transport) *ISteamDirectory {
	// The transport is validated here in case this package is used directly, outside of the API wrapper which validates and returns an error.
	// Panicking is intentional since this runs during lazy service initialization, where propagating errors would decrease usability.
	if transport == nil {
		panic("transport is nil")
	}
	return &ISteamDirectory{transport: transport}
}
