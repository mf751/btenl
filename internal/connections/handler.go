package connections

import (
	"context"

	"github.com/mf751/btenl/internal/shared/types"
)

func (c *ConnectionMux) handleConn(
	ctx context.Context,
	conn types.Conn,
	sink types.ConnectionSink,
) {
	// INFO: merged context: daemon ctx + conn ctx
	mCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(conn.Ctx(), cancel)
	defer stop()

	for {
		select {
		case <-mCtx.Done():
			return
		default:
		}
		stream, err := conn.AcceptStream(mCtx)
		if err != nil {
			ev := types.ConnectionEvent{
				Event:  types.NewEvent(types.EventTypeConnectionError),
				ConnID: conn.ID(),
			}
			sink.Connections(ev)
			continue
		}

		go handleStream(mCtx, stream, conn, sink)
	}
}

func (c *ConnectionMux) handleConnErr(
	ctx context.Context,
	err error,
	sink types.ConnectionSink,
) {
	ev := types.NewEvent(types.EventTypeConnectionError)
	sink.Errors(
		types.ErrorEvent{Err: err, Event: ev},
	)
}

func handleStream(
	ctx context.Context,
	stream types.Stream,
	conn types.Conn,
	sink types.ConnectionSink,
) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	ev := types.ConnectionEvent{
		Event:  types.NewEvent(types.EventTypeStreamOpened),
		ConnID: conn.ID(),
		Stream: stream,
	}

	// forward upstream to daemon
	sink.Connections(ev)
}
