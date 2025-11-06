package connection

import (
	"context"
	"net"
)

// Dialer defines the interface for creating network connections.
//
//go:generate mockgen -package=connection_test -destination=mock_dialer_test.go github.com/lewisgibson/go-steam/connection Dialer
//go:generate mockgen -package=connection_test -destination=mock_net_conn_test.go net Conn
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
