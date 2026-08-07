package daemon

import (
	"fmt"
	"time"

	"github.com/mf751/btenl.git/internal/logger"
	"github.com/mf751/btenl.git/internal/types"
)

const (
	RESET   string = "\033[0m"
	RED     string = "\033[31m"
	GREEN   string = "\033[32m"
	BLUE    string = "\033[34m"
	MAGENTA string = "\033[35m"
	CYAN    string = "\033[36m"
)

func prettierDate(t time.Time) string {
	return t.Format("2006/01/02 15:04:05")
}

func eventTypeString(t types.EventType) string {
	switch t {
	case types.EventTypeControlError:
		return "CONTROL"
	default:
		return "EVENT ERROR"
	}
}

func printErrorEvent(event types.ErrorEvent) {
	fmt.Print("[" + RED + "EVENTERROR" + RESET + "]")
	fmt.Print("[" + BLUE + "DATE" + RESET + "]")
	fmt.Print("[" + RED + prettierDate(event.StartedAt) + RESET + "]")
	fmt.Print("[" + BLUE + "TYPE" + RESET + "]")
	fmt.Print("[" + RED + eventTypeString(event.Type) + RESET + "]")
	fmt.Print("[" + BLUE + "MESSAGE" + RESET + "]")
	fmt.Print("[" + RED + event.Err.Error() + RESET + "]")
	fmt.Println()
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
	printErrorEvent(event)
	logErrorEvent(d.logger, event)
}
