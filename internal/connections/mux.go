package connections

import (
	"context"
	"sync"

	"github.com/mf751/btenl.git/internal/shared/types"
)

// connectionMux implements ConnectionsEventSource and receivs multiple
// connection sources and manages connections lifecycle and forwards to
// the daemon ConnectionEvents which has node field that represnets the
// connection.
// it also implements the ConnectionManager interface which the daemon
// uses to access connections and perform actions on connections
type ConnectionMux struct {
	connectors []Connector

	conns map[types.ConnID][]types.Conn

	mu sync.RWMutex
}

func NewConnectionMux(ctrs ...Connector) *ConnectionMux {
	return &ConnectionMux{
		connectors: ctrs,
		conns:      make(map[types.ConnID][]types.Conn),
	}
}

func (c *ConnectionMux) ServeEvents(ctx context.Context, sink types.ConnectionSink) error {
	onConn := func(conn types.Conn) {
		c.mu.Lock()
		c.conns[conn.ID()] = append(c.conns[conn.ID()], conn)
		c.mu.Unlock()
	}

	onErr := func(err error) {
		ev := types.NewEvent(types.EventTypeConnectionError)
		sink.Errors(
			types.ErrorEvent{Err: err, Event: ev},
		)
	}

	for _, ctr := range c.connectors {
		go func() {
			err := ctr.Listen(ctx, onConn, onErr)
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
