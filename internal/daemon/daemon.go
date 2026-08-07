package daemon

import (
	"context"

	"github.com/mf751/btenl.git/internal/types"
)

type Daemon struct {
	ctx  context.Context
	kill context.CancelFunc

	errorEvents      chan types.ErrorEvent
	controlEvents    chan types.ControlEvent
	connectionEvents chan types.ConnectionEvent
}

func New(ctx context.Context) *Daemon {
	ctx, cancel := context.WithCancel(ctx)
	return &Daemon{
		ctx:              ctx,
		kill:             cancel,
		errorEvents:      make(chan types.ErrorEvent, 128),
		controlEvents:    make(chan types.ControlEvent, 128),
		connectionEvents: make(chan types.ConnectionEvent, 128),
	}
}

func (d *Daemon) Controls(event types.ControlEvent) {
	select {
	case d.controlEvents <- event:
	case <-d.ctx.Done():
	}
}

func (d *Daemon) Errors(event types.ErrorEvent) {
	select {
	case d.errorEvents <- event:
	case <-d.ctx.Done():
	}
}

func (d *Daemon) ForwardControlSource(src types.ControlEventSource) {
	// INFO: don't run if context is done
	select {
	case <-d.ctx.Done():
		return
	default:
	}
	go func() {
		err := src.ServeEvents(d.ctx, d)
		// NOTE: runs only when the source stops serving
		if err != nil && d.ctx.Err() == nil {
			d.Errors(types.ErrorEvent{Err: err})
		}
	}()
}

func (d *Daemon) worker() {
	for {
		select {
		case event := <-d.controlEvents:
			d.handleControlEvent(event)
		case event := <-d.connectionEvents:
			d.handleConnectionEvent(event)
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Daemon) errorWorker() {
	for {
		select {
		case event := <-d.errorEvents:
			d.handleErrorEvent(event)
		case <-d.ctx.Done():
			return
		}
	}
}

// INFO: Daemon Event Loop
func (d *Daemon) Run() {
	const eventWorkers int = 4
	for range eventWorkers {
		go d.worker()
	}
	go d.errorWorker()
	<-d.ctx.Done()
}
