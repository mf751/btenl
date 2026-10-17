package connections

import (
	"context"
	"errors"
	"net"

	"github.com/mf751/btenl/internal/shared/types"
)

// sink must be set before use
func (c *ConnectionMux) SetSink(sink types.ConnectionSink) {
	c.sink = sink
}

func (c *ConnectionMux) Connect(ctx context.Context, ip net.IP) (types.ConnID, error) {
	// TODO: fallthrough connectors when failing until a successfull connections or return err
	conn, err := c.connectors[0].Connect(ctx, ip)
	if err != nil {
		c.handleConnErr(ctx, err, c.sink)
		return 0, err
	}
	c.state.AddConn(conn)

	go c.handleConn(ctx, conn, c.sink)

	return conn.ID(), nil
}

func (c *ConnectionMux) OpenStream(ctx context.Context, connID types.ConnID) (types.Stream, error) {
	conn, ok := c.state.GetConn(connID)
	if !ok {
		return nil, errors.New("no connection with given id")
	}

	stream, err := conn.OpenStream(ctx)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *ConnectionMux) GetConns(ctx context.Context) []types.ConnInfo {
	return c.state.GetAllConns()
}
