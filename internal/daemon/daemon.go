package daemon

import (
	"context"

	"github.com/mf751/btenl.git/internal/types"
)

type Daemon struct {
	events chan types.Event
	ctx    context.Context
	kill   context.CancelFunc
}

func New(ctx context.Context, sources ...types.EventSource) *Daemon {
	ctx, cancel := context.WithCancel(ctx)
	d := &Daemon{
		events: make(chan types.Event, 128),
		ctx:    ctx,
		kill:   cancel,
	}

	for _, src := range sources {
		d.ForwardSource(src)
	}

	return d
}

func (d *Daemon) ForwardSource(src types.EventSource) {
	// INFO: don't run if context is done
	select {
	case <-d.ctx.Done():
		return
	default:
	}
	go func() {
		err := src.ServeEvents(d.ctx, d.events)
		// NOTE: runs only when the source stops serving
		if err != nil {
			select {
			case d.events <- types.ErrorEvent{Err: err}:
			case <-d.ctx.Done():
				return
			}
		}
	}()
}

// INFO: Daemon Event Loop
func (d *Daemon) Run() {
	for {
		select {
		case ev := <-d.events:
			switch e := ev.(type) {
			case types.ControlEvent:
				d.handleControlEvent(e)
			case types.ConnectionEvent:
				d.handleConnectionEvent(e)
			case types.ErrorEvent:
				d.handleErrorEvent(e)
			}
		case <-d.ctx.Done():
			return
		}
	}
}
