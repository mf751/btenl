package daemon

import "github.com/mf751/btenl.git/internal/types"

func (d *Daemon) handleControlEvent(event types.ControlEvent) {
	event.Reply(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
}
