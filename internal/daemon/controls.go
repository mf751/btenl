package daemon

import (
	"time"

	"github.com/mf751/btenl.git/internal/types"
)

func (d *Daemon) handleControlEvent(event types.ControlEvent) {
	event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
	time.Sleep(3 * time.Second)
	event.Send(types.ControlResponse{Status: types.StatusSucceed, Msg: "Hello"})
	event.Close()
}
