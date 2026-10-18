package connections

import (
	"sync"

	"github.com/mf751/btenl/internal/shared/types"
)

// State holds the runtime state of the ConnectionMux: every connection
// it guards its maps with a Read Write mutex and
// is safe for concurrent use by the mux goroutines.
type State struct {
	mu    sync.RWMutex
	conns map[types.ConnID]types.Conn
}

func NewState() *State {
	return &State{
		conns: make(map[types.ConnID]types.Conn),
	}
}

// AddConn registers a connection under its id. a connection can appear
// multiple times under the same id when there are multiple transport
// paths to the same peer.
func (s *State) AddConn(conn types.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.conns[conn.ID()] = conn
}

func (s *State) GetConn(connID types.ConnID) (types.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, ok := s.conns[connID]
	return conn, ok
}

func (s *State) GetAllConns() []types.ConnInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := []types.ConnInfo{}
	for _, conn := range s.conns {
		conns = append(conns, types.ConnInfo{ID: conn.ID(), NodeID: conn.NodeID(), IP: conn.IP()})
	}
	return conns
}

func (s *State) DeleteConn(connID types.ConnID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, connID)
}
