package daemon

import (
	"time"

	"github.com/mf751/btenl.git/internal/shared/logger"
	"github.com/mf751/btenl.git/internal/shared/types"
)

func eventTypeString(t types.EventType) string {
	switch t {
	case types.EventTypeControlError:
		return "CONTROL"
	default:
		return "EVENT ERROR"
	}
}

func logErrorEvent(logger *logger.Logger, event types.ErrorEvent) {
	logger.Errorf(
		"EventStarted:%s Type:%s Message:%s",
		event.StartedAt.Format(time.RFC3339),
		eventTypeString(event.Type),
		event.Err.Error(),
	)
}

func (d *Daemon) handleErrorEvent(event types.ErrorEvent) {
	logErrorEvent(d.logger, event)
}
