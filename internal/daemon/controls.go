package daemon

import (
	"time"

	"github.com/mf751/btenl.git/internal/types"
)

func (d *Daemon) handleControlEvent(event types.ControlEvent) {
	switch event.Command {
	case "stop":
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "stopping"})
		event.Close()
		// NOTE: give the response time to flush before shutting down
		time.AfterFunc(200*time.Millisecond, d.Stop)
	default:
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		time.Sleep(3 * time.Second)
		event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
		event.Close()
	}
}
