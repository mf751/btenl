package types

import "context"

type Event interface {
	isEvent()
}

type ErrorEvent struct {
	Err error
}

func (ErrorEvent) isEvent() {}

type EventSource interface {
	Serve(ctx context.Context, out chan<- Event) error
}
