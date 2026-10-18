package daemon

import (
	"context"
	"net"

	"github.com/mf751/btenl/internal/shared/types"
)

// ConnectionManager
// represents the session layer in the communication stack manages
// multiple connectors and connections nodes and thier lifecycles and provides
// methods to initiate connections or perform actions on connected nodes
type ConnectionManager interface {
	Connect(ctx context.Context, ip net.IP) (types.ConnID, error)
	Disconnect(connID types.ConnID) error
	OpenStream(ctx context.Context, connID types.ConnID) (types.Stream, error)
	GetConns() []types.ConnInfo
}

func (d *Daemon) handleConnectionEvent(event types.ConnectionEvent) {
	buff := make([]byte, 20)

	if event.Type != types.EventTypeConnectionError &&
		event.Type != types.EventTypeConnectionClosed {
		n, err := event.Stream.Read(buff)
		if err != nil {
			return
		}
		if string(buff[:n]) == "Ping!" {
			event.Stream.Write([]byte("Pong!"))
			event.Stream.Close()
		}
	}
}
