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
	Connect(ctx context.Context, ip net.IP) (*types.Node, error)
	ListNodes() []*types.Node
}

func (d *Daemon) handleConnectionEvent(event types.ConnectionEvent) {
}
