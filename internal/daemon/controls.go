package daemon

import (
	"fmt"
	"net"
	"time"

	"github.com/mf751/btenl/internal/shared/types"
)

func (d *Daemon) handleControlEvent(event types.ControlEvent) {
	switch event.Command {

	case "kill":
		event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: "killing daemon"})
		event.Close()
		// NOTE: give the response time to flush before shutting down
		time.AfterFunc(200*time.Millisecond, d.Kill)

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
				Status: types.StatusSucceeded,
				Msg:    fmt.Sprintf("connected to %x", connID),
			},
		)
		event.Close()

	case "disconnect":
		err := d.manager.Disconnect(d.manager.GetConns()[0].ID)
		if err != nil {
			event.Send(types.ControlResponse{Status: types.StatusFailed, Msg: err.Error()})
			event.Close()
			return
		}
		event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: "Closed connection"})
		event.Close()

	case "conns":
		conns := d.manager.GetConns()
		s := "Open Connections: \n"
		for _, connInfo := range conns {
			s += fmt.Sprintf(
				"conn-%d --> %v :- %v\n",
				connInfo.ID,
				connInfo.NodeID.ShortString(),
				connInfo.IP.String(),
			)
		}
		event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: s})
		event.Close()

	case "ping":
		stream, err := d.manager.OpenStream(d.ctx, d.manager.GetConns()[0].ID)
		if err != nil {
			event.Send(
				types.ControlResponse{
					Status: types.StatusTimout,
					Msg:    "failed to open stream with connection" + string(types.ConnID(1)),
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
					Msg:    "failed to write to stream with connection" + string(types.ConnID(1)),
				},
			)
			event.Close()
			return
		}
		event.Send(
			types.ControlResponse{
				Status: types.StatusSucceeded,
				Msg:    "Waiting for pong!" + string(types.ConnID(1)),
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
			event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: "Pong!"})
			event.Close()
			return
		}
		event.Send(types.ControlResponse{Status: types.StatusInvalidResponse, Msg: string(res)})
		event.Close()

	default:
		event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: "Hello"})
		time.Sleep(3 * time.Second)
		event.Send(types.ControlResponse{Status: types.StatusSucceeded, Msg: "Hello"})
		event.Close()
	}
}
