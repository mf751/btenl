//go:build unix

package control

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"

	"github.com/mf751/btenl.git/internal/types"
)

type UnixIPCSource struct {
	socketPath string
}

func NewUnixIPCSource(socketPath string) *UnixIPCSource {
	return &UnixIPCSource{
		socketPath: socketPath,
	}
}

type ControlConn struct {
	conn   net.Conn
	sink   types.ControlSink
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *ControlConn) log(err error) {
	c.sink.Errors(types.ErrorEvent{
		Err:   err,
		Event: types.NewEvent(types.EventTypeControlError),
	})
}

func (c *ControlConn) send(res types.ControlResponse) {
	err := json.NewEncoder(c.conn).Encode(&res)
	if err != nil {
		c.log(err)
	}
}

func (u *UnixIPCSource) ServeEvents(ctx context.Context, sink types.ControlSink) error {
	// NOTE: remove old socket if exists
	if err := os.RemoveAll(u.socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", u.socketPath)
	if err != nil {
		return err
	}

	defer listener.Close()
	defer os.Remove(u.socketPath)

	// NOTE: stop listener when ctx done
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// NOTE: since listener.Accept blocks check ctx done after acceptinng
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}

		connCtx, cancel := context.WithCancel(ctx)
		c := &ControlConn{conn: conn, ctx: connCtx, sink: sink, cancel: cancel}

		go u.handleConnection(c)
	}
}

func (u *UnixIPCSource) handleConnection(c *ControlConn) {
	// NOTE: closes the connection and ends ctx of c when this functoin is exits
	defer func() {
		c.conn.Close()
		c.cancel()
	}()

	var req types.ControlRequest

	if err := json.NewDecoder(c.conn).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return
		}

		c.send(types.ControlResponse{Status: types.StatusInvalidRequest, Msg: "invalid request"})
		c.log(err)
		return
	}

	ch := make(chan types.ControlResponse, 16)

	event := types.ControlEvent{
		Event:   types.NewEvent(types.EventTypeControlCommand),
		Command: req.Command,
		Args:    req.Args,
		Send: func(res types.ControlResponse) error {
			select {
			case ch <- res:
				return nil
			case <-c.ctx.Done():
				return c.ctx.Err()
			}
		},
		Close: func() {
			close(ch)
		},
	}

	// NOTE: send control event to daemon
	c.sink.Controls(event)
	select {
	case <-c.ctx.Done():
		return
	default:
	}

	// NOTE: wait for responses from daemon until c is ended by daemon
	for {
		select {
		case res, ok := <-ch:
			if !ok {
				return
			}
			if err := json.NewEncoder(c.conn).Encode(res); err != nil {
				c.log(err)
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}
