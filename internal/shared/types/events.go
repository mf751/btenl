package types

import (
	"context"
	"time"
)

type EventType uint8

const (
	EventTypeControlCommand EventType = iota
	EventTypeControlError

	EventTypeConnectionNew
	EventTypeConnectionData
	EventTypeConnectionClosed
	EventTypeConnectionError

	EventTypeError
)

// NOTE: All Events contain this struct
type Event struct {
	StartedAt time.Time
	Type      EventType
}

func NewEvent(eType EventType) Event {
	return Event{
		StartedAt: time.Now(),
		Type:      eType,
	}
}

type EventSource[S Sink] interface {
	ServeEvents(ctx context.Context, sink S) error
}

type (
	ControlEventSource    = EventSource[ControlSink]
	ConnectionEventSource = EventSource[ConnectionSink]
)

//

//

// INFO: Error
type ErrorEvent struct {
	Event
	Err error
}

//

//

// INFO: Controls layer
type ControlEvent struct {
	Event
	Command string
	Args    []string
	Send    func(ControlResponse) error
	Close   func()
}

//

//

// INFO: Connections layer
type ConnectionEvent struct {
	Event
	// Node Node
	// Data any
	// send()
	// close()
}
