// Daemon (main event loops)
// it can receive multiple EventSources( control / connections ) via daemon.ForwardSource functions
// it is the highest layer in the communication stack receives pure Event types has no knowledge
// of the underlying communication methods used
// implements all Sink interfaces so it can be passed to all EventSource types that need Sink methods
// to send events to and it forwards them into its event channels
// on daemon.Run it creates worker pool that read from the channels and handle the events

package daemon

import (
	"context"
	"runtime"

	"github.com/mf751/btenl.git/internal/shared/logger"
	"github.com/mf751/btenl.git/internal/shared/types"
)

type Daemon struct {
	ctx  context.Context
	kill context.CancelFunc

	logger  *logger.Logger
	manager ConnectionManager

	errorEvents      chan types.ErrorEvent
	controlEvents    chan types.ControlEvent
	connectionEvents chan types.ConnectionEvent
}

func New(ctx context.Context, manager ConnectionManager, logger *logger.Logger) *Daemon {
	ctx, cancel := context.WithCancel(ctx)
	return &Daemon{
		ctx:              ctx,
		kill:             cancel,
		logger:           logger,
		manager:          manager,
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

func (d *Daemon) Connections(event types.ConnectionEvent) {
	select {
	case d.connectionEvents <- event:
	case <-d.ctx.Done():
	}
}

func (d *Daemon) Errors(event types.ErrorEvent) {
	select {
	case d.errorEvents <- event:
	case <-d.ctx.Done():
	}
}

func (d *Daemon) Stop() {
	d.kill()
}

func (d *Daemon) eventWorker() {
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
	eventWorkers := min(runtime.NumCPU(), 6)
	for range eventWorkers {
		go d.eventWorker()
	}
	go d.errorWorker()

	d.logger.Infof("Daemon Started (%d workers)", eventWorkers)

	// NOTE: keep running until ctx is done
	<-d.ctx.Done()
	d.logger.Info("Daemon Stopped")
}
