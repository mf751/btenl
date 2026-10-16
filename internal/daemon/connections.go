package daemon

import (
	"context"
	"net"

	"github.com/mf751/btenl.git/internal/shared/types"
)

// ConnectionManager
// represents the session layer in the communication stack manages
// multiple connectors and connections nodes and thier lifecycles and provides
// methods to initiate connections or perform actions on connected nodes
type ConnectionManager interface {
	Connect(ctx context.Context, ip net.IP) (types.ConnID, error)
	OpenStream(ctx context.Context, connID types.ConnID) (types.Stream, error)
	GetConns(ctx context.Context) []types.ConnID
}

func (d *Daemon) handleConnectionEvent(event types.ConnectionEvent) {
	buff := make([]byte, 20)

	if event.Type != types.EventTypeConnectionError {
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
