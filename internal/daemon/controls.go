package daemon

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/mf751/btenl.git/internal/shared/types"
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

		node, err := d.manager.Connect(d.ctx, ip)
		if err != nil {
			event.Send(types.ControlResponse{Status: types.StatusDaemonError, Msg: err.Error()})
			event.Close()
			return
		}

		event.Send(
			types.ControlResponse{
				Status: types.StatusSucceed,
				Msg:    fmt.Sprintf("connected to %x", node.ID),
			},
		)
		event.Close()

	case "nodes":
		nodes := d.manager.ListNodes()
		msg := "Connected Nodes:\n\n"
		for i, n := range nodes {
			msg += fmt.Sprintf("%v - %v\n", i+1, hex.EncodeToString(n.ID[:]))
		}
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: msg})
		event.Close()

	default:
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		time.Sleep(3 * time.Second)
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		event.Close()
	}
}
