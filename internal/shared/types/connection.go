package types

import (
	"context"
	"io"
)

// Represents and open secure tunnel in a connection
// that can be written to and read from and closed by
// both ends of the connection
type Stream interface {
	io.Reader
	io.Writer
	io.Closer
}

type ConnID [32]byte

// Represents a peer-to-peer connection between two btenl
// instances and is kept alive until one peer closes it
// a connection can have multiple independent bidirectional
// streams that are used to transfer data between peers
type Conn interface {
	ID() ConnID
	OpenStream(ctx context.Context) (Stream, error)
	AcceptStream(ctx context.Context) (Stream, error)
	Close() error
}

type Node struct {
	ID ConnID
}
