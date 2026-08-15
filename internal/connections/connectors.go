package connections

import (
	"context"
	"net"

	"github.com/mf751/btenl.git/internal/shared/types"
)

// Connector
// Implemented by any connection provider in the transport layer
// quic, tcp, any other protocols that accepts connections and calls onConn
// and provides Connect to dial connections on given ip
// it calls onErr on any error that doesn't end the listener which then the
// connection mux forwards to the daemon error eventloop
// returns error when the listener stops
type Connector interface {
	Listen(ctx context.Context, onConn func(types.Conn), onErr func(error)) error
	Connect(ctx context.Context, ip net.IP) (types.Conn, error)
}

// the order of adding connectors matters because when calling
// ConnectionMux.Connect it will try to use the first added Connector
// in the slice and keep falling back to the next Connector if fails
func (c *ConnectionMux) AddConnectors(ctrs ...Connector) {
	c.connectors = append(c.connectors, ctrs...)
}
