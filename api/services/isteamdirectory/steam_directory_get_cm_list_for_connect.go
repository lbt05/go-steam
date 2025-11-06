package isteamdirectory

import (
	"context"
	"fmt"
	"net/http"
)

// GetCMListForConnectParameters describes the parameters for the GetCMListForConnect method.
type GetCMListForConnectParameters struct {
	CellID   uint32 `url:"cellid,omitempty"`
	CMType   string `url:"cmtype,omitempty"`
	Realm    string `url:"realm,omitempty"`
	MaxCount uint32 `url:"maxcount,omitempty"`
}

// GetCMListForConnectResponse describes the response for the GetCMListForConnect method.
type GetCMListForConnectResponse struct {
	Response struct {
		Success    bool       `json:"success"`
		Message    string     `json:"message"`
		ServerList []CMServer `json:"serverlist"`
	} `json:"response"`
}

// CMServer describes the server for the GetCMListForConnect method.
type CMServer struct {
	Endpoint       string  `json:"endpoint,omitempty"`
	LegacyEndpoint string  `json:"legacy_endpoint,omitempty"`
	Type           string  `json:"type,omitempty"`
	DC             string  `json:"dc,omitempty"`
	Realm          string  `json:"realm,omitempty"`
	Load           int     `json:"load,omitempty"`
	WtdLoad        float64 `json:"wtd_load,omitempty"`
}

// GetCMListForConnect returns a list of CM servers to connect to.
func (a *ISteamDirectory) GetCMListForConnect(ctx context.Context, params GetCMListForConnectParameters) ([]CMServer, error) {
	var res GetCMListForConnectResponse
	if err := a.transport.Call(ctx, http.MethodGet, "ISteamDirectory", "GetCMListForConnect", 1, params, &res); err != nil {
		return nil, err
	}
	if !res.Response.Success {
		return nil, fmt.Errorf("failed to get CM list for connect: %s", res.Response.Message)
	}
	return res.Response.ServerList, nil
}
