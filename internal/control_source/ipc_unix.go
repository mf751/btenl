//go:build unix

package control

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"time"

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
	out    chan<- types.Event
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *ControlConn) log(err error) {
	error := types.ErrorEvent{
		Err: err,
	}
	select {
	case c.out <- error:
	case <-c.ctx.Done():
		return
	}
}

func (c *ControlConn) send(res types.ControlResponse) {
	err := json.NewEncoder(c.conn).Encode(&res)
	if err != nil {
		c.log(err)
	}
}

func (u *UnixIPCSource) Serve(ctx context.Context, out chan<- types.Event) error {
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
		c := &ControlConn{conn: conn, ctx: connCtx, out: out, cancel: cancel}

		go u.handleConnection(c)
	}
}

func (u *UnixIPCSource) handleConnection(c *ControlConn) {
	defer func() {
		c.conn.Close()
		c.cancel()
	}()

	// NOTE: close conn when ctx done
	go func() {
		<-c.ctx.Done()
		c.conn.Close()
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

	reply := make(chan types.ControlResponse, 1)

	event := types.ControlEvent{
		Command: req.Command,
		Args:    req.Args,
		Reply: func(res types.ControlResponse) error {
			select {
			case reply <- res:
				return nil
			case <-c.ctx.Done():
				return c.ctx.Err()
			}
		},
	}

	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()

	// NOTE: wait for either writing to out or ctx done or 30s
	select {
	case c.out <- event:
	case <-timer.C:
		c.send(types.ControlResponse{Status: types.StatusTimout, Msg: "connection timed out"})
		return
	case <-c.ctx.Done():
		c.send(types.ControlResponse{Status: types.StatusDaemonDied, Msg: "daemon died"})
		return
	}

	// NOTE: wait for either res written to or ctx done or 30s
	select {
	case res := <-reply:
		if err := json.NewEncoder(c.conn).Encode(res); err != nil {
			c.log(err)
			return
		}
	case <-timer.C:
		c.send(types.ControlResponse{Status: types.StatusTimout, Msg: "connection timed out"})
	case <-c.ctx.Done():
		c.send(types.ControlResponse{Status: types.StatusDaemonDied, Msg: "daemon died"})
		return
	}
}
