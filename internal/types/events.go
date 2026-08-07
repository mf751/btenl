package types

import "context"

// NOTE: all events must implement this interface
type Event interface {
	isEvent()
}

type ControlEventSource interface {
	ServeEvents(ctx context.Context, sink ControlSink) error
}

//

//

// INFO: Error
type ErrorEvent struct {
	Err error
}

func (ErrorEvent) isEvent() {}

//

//

// INFO: Controls layer
type ControlEvent struct {
	Command string
	Args    []string

	Send  func(ControlResponse) error
	Close func()
}

func (ControlEvent) isEvent() {}

//

//

// INFO: Connections layer
type ConnectionEvent struct {
	// Node Node
	// Data any
	// send()
	// close()
}

func (ConnectionEvent) isEvent() {}
