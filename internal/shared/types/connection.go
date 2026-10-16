package types

import (
	"context"
	"encoding/hex"
	"io"
	"sync/atomic"
)

// Represents an open secure tunnel in a connection
// that can be written to and read from and closed by
// both ends of the connection
type Stream interface {
	io.Reader
	io.Writer
	io.Closer
}

const (
	StreamTypePing uint8 = iota
	StreamTypeFile
)

// Unique per connection on every run
type ConnID uint64

// atomic Uint64 not to be copied
var connSeq atomic.Uint64

// increasing integer for every new connection
func NextConnID() uint64 { return connSeq.Add(1) }

type NodeID [32]byte

func (c NodeID) String() string {
	return hex.EncodeToString(c[:])
}

// Represents a peer-to-peer connection between two btenl
// instances and is kept alive until one peer closes it
// a connection can have multiple independent bidirectional
// streams that are used to transfer data between peers
type Conn interface {
	ID() ConnID
	NodeID() NodeID
	Ctx() context.Context
	Close() error
	OpenStream(ctx context.Context) (Stream, error)
	AcceptStream(ctx context.Context) (Stream, error)
}
