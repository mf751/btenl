package daemon

import (
	"context"

	"github.com/mf751/btenl.git/internal/shared/types"
)

// EventSource has 2 types:
//
//   - ControlEventSource: implemented by all Control Event Providers
//     that accept a ControlSink to forward Control and Error Events to the sink
//     and has this signature:
//     type ControlEventSource interface {
//         ServeEvents(ctx context.Context, sink ControlSink) error
//     }
//
//   - ConnectionSinkEventSource: implemented by all Connection Event Providers
//     that accept a ConnectionSink to forward Connection and Error Events to the sink
//     and has this signature:
//     type ConnectionEventSource interface {
//         ServeEvents(ctx context.Context, sink ConnectionSink) error
//     }

type EventSource[S types.Sink] interface {
	ServeEvents(ctx context.Context, sink S) error
}

type (
	ControlEventSource    = EventSource[types.ControlSink]
	ConnectionEventSource = EventSource[types.ConnectionSink]
)

func forward[S types.Sink](d *Daemon, src EventSource[S], sink S, et types.EventType) {
	select {
	case <-d.ctx.Done():
		return
	default:
	}

	go func() {
		err := src.ServeEvents(d.ctx, sink)
		if err != nil && d.ctx.Err() == nil {
			d.Errors(types.ErrorEvent{Err: err, Event: types.NewEvent(et)})
		}
	}()
}

func (d *Daemon) ForwardControlSource(src ControlEventSource) {
	forward[types.ControlSink](d, src, d, types.EventTypeControlError)
}

func (d *Daemon) ForwardConnectionSource(src ConnectionEventSource) {
	forward[types.ConnectionSink](d, src, d, types.EventTypeConnectionError)
}
