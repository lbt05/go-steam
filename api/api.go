package api

import (
	"fmt"

	"github.com/lbt05/go-steam/api/services/iauthenticationservice"
	"github.com/lbt05/go-steam/api/services/ieconservice"
	"github.com/lbt05/go-steam/api/services/isteamdirectory"
	"github.com/lbt05/go-steam/api/services/itwofactorservice"
	"github.com/lbt05/go-steam/api/transports"
	"github.com/lbt05/go-steam/internal/lazy"
)

// ErrNilTransport is the error returned when the transport is nil.
var ErrNilTransport = fmt.Errorf("transport is nil")

// API is the main API client.
type API struct {
	iEconService           *lazy.Value[*ieconservice.IEconService]
	iSteamDirectory        *lazy.Value[*isteamdirectory.ISteamDirectory]
	iAuthenticationService *lazy.Value[*iauthenticationservice.IAuthenticationService]
	iTwoFactorService      *lazy.Value[*itwofactorservice.ITwoFactorService]
}

// NewAPI creates a new API instance with the given transport.
func NewAPI(transport transports.Transport) (*API, error) {
	if transport == nil {
		return nil, ErrNilTransport
	}
	return &API{
		iEconService: lazy.New(func() *ieconservice.IEconService {
			return ieconservice.NewIEconService(transport)
		}),
		iSteamDirectory: lazy.New(func() *isteamdirectory.ISteamDirectory {
			return isteamdirectory.NewISteamDirectory(transport)
		}),
		iAuthenticationService: lazy.New(func() *iauthenticationservice.IAuthenticationService {
			return iauthenticationservice.NewIAuthenticationService(transport)
		}),
		iTwoFactorService: lazy.New(func() *itwofactorservice.ITwoFactorService {
			return itwofactorservice.NewITwoFactorService(transport)
		}),
	}, nil
}

// IEconService returns the IEconService service.
func (s *API) IEconService() *ieconservice.IEconService {
	return s.iEconService.Get()
}

// SteamDirectory returns the SteamDirectory service.
func (s *API) SteamDirectory() *isteamdirectory.ISteamDirectory {
	return s.iSteamDirectory.Get()
}

// IAuthenticationService returns the IAuthenticationService service.
func (s *API) IAuthenticationService() *iauthenticationservice.IAuthenticationService {
	return s.iAuthenticationService.Get()
}

// ITwoFactorService returns the ITwoFactorService service.
func (s *API) ITwoFactorService() *itwofactorservice.ITwoFactorService {
	return s.iTwoFactorService.Get()
}
