package connections

import (
	"context"

	"github.com/mf751/btenl/internal/shared/types"
)

// ConnectionMux
// implements ConnectionsEventSource and receives multiple
// connection sources and manages connections lifecycle and forwards to
// the daemon ConnectionEvents which has node field that represents the
// connection.
// it also implements the ConnectionManager interface which the daemon
// uses to access connections and perform actions on connections
type ConnectionMux struct {
	connectors []Connector
	sink       types.ConnectionSink

	state *State
}

func NewConnectionMux(ctrs ...Connector) *ConnectionMux {
	return &ConnectionMux{
		connectors: ctrs,
		state:      NewState(),
	}
}

func (c *ConnectionMux) ServeEvents(ctx context.Context, sink types.ConnectionSink) error {
	for _, ctr := range c.connectors {
		go func() {
			err := ctr.Listen(
				ctx,
				func(conn types.Conn) {
					c.state.AddConn(conn)
					c.handleConn(ctx, conn, sink)
				},
				func(err error) { c.handleConnErr(ctx, err, sink) },
			)
			if err != nil {
				sink.Errors(
					types.ErrorEvent{
						Event: types.NewEvent(types.EventTypeConnectionError),
						Err:   err,
					},
				)
			}
		}()
	}

	<-ctx.Done()
	return nil
}
