package types

// Sink represents an event receiver
type Sink interface {
	Errors(ErrorEvent)
}

type ControlSink interface {
	Controls(ControlEvent)
	Sink
}

type ConnectionSink interface {
	Connections(ConnectionEvent)
	Sink
}
