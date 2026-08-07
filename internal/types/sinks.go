package types

type ErrorSink interface {
	Errors(ErrorEvent)
}

type ControlSink interface {
	Controls(ControlEvent)
	ErrorSink
}
