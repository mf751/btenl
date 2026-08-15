package connections

import (
	"context"
	"net"

	"github.com/mf751/btenl.git/internal/shared/types"
)

func (c *ConnectionMux) Connect(ctx context.Context, ip net.IP) (*types.Node, error) {
	conn, err := c.connectors[0].Connect(ctx, ip)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.conns[conn.ID()] = append(c.conns[conn.ID()], conn)
	c.mu.Unlock()

	return &types.Node{ID: conn.ID()}, nil
}

func (c *ConnectionMux) ListNodes() []*types.Node {
	nodes := []*types.Node{}
	c.mu.Lock()
	for key := range c.conns {
		nodes = append(nodes, &types.Node{ID: key})
	}
	c.mu.Unlock()
	return nodes
}
