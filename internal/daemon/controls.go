package daemon

import (
	"fmt"
	"net"
	"time"

	"github.com/mf751/btenl/internal/shared/types"
)

func (d *Daemon) handleControlEvent(event types.ControlEvent) {
	switch event.Command {

	case "stop":
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "stopping"})
		event.Close()
		// NOTE: give the response time to flush before shutting down
		time.AfterFunc(200*time.Millisecond, d.Stop)

	case "connect":
		if len(event.Args) < 1 {
			event.Send(
				types.ControlResponse{
					Status: types.StatusInvalidRequest,
					Msg:    "must provide target ip",
				},
			)
			event.Close()
			return
		}
		ip := net.ParseIP(event.Args[0])
		if ip == nil {
			event.Send(types.ControlResponse{Status: types.StatusDaemonError, Msg: "invalid ip"})
			event.Close()
			return
		}

		connID, err := d.manager.Connect(d.ctx, ip)
		if err != nil {
			event.Send(types.ControlResponse{Status: types.StatusDaemonError, Msg: err.Error()})
			event.Close()
			return
		}

		event.Send(
			types.ControlResponse{
				Status: types.StatusSucceed,
				Msg:    fmt.Sprintf("connected to %x", connID),
			},
		)
		event.Close()

	case "connections":
		conns := d.manager.GetConns(d.ctx)
		s := "Open Connections: \n"
		for k := range conns {
			s += fmt.Sprintf("- %d", k)
		}
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: s})
		event.Close()

	case "ping":
		conns := d.manager.GetConns(d.ctx)
		stream, err := d.manager.OpenStream(d.ctx, conns[0])
		if err != nil {
			event.Send(
				types.ControlResponse{
					Status: types.StatusTimout,
					Msg:    "failed to open stream with connection" + string(conns[0]),
				},
			)
			event.Close()
			return
		}
		_, err = stream.Write([]byte("Ping!"))
		if err != nil {
			event.Send(
				types.ControlResponse{
					Status: types.StatusTimout,
					Msg:    "failed to write to stream with connection" + string(conns[0]),
				},
			)
			event.Close()
			return
		}
		event.Send(
			types.ControlResponse{
				Status: types.StatusSucceed,
				Msg:    "Waiting for pong!" + string(conns[0]),
			},
		)
		res := make([]byte, 5)
		n, err := stream.Read(res)
		// NOTE: a Read may return n > 0 together with io.EOF (the peer wrote
		// "Pong!" and closed its send side), so only fail when nothing was read.
		if err != nil && n == 0 {
			event.Send(
				types.ControlResponse{Status: types.StatusTimout, Msg: "failed to read response "},
			)
			event.Close()
			return
		}
		if string(res[:n]) == "Pong!" {
			event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Pong!"})
			event.Close()
			return
		}
		event.Send(types.ControlResponse{Status: types.StatusInvalidResponse, Msg: string(res)})
		event.Close()

	default:
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		time.Sleep(3 * time.Second)
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		event.Close()
	}
}
