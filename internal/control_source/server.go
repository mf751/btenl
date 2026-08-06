package control

import (
	"context"
	"sync"

	"github.com/mf751/btenl.git/internal/types"
)

type ControlSource struct {
	events  chan types.Event
	sources []types.EventSource
	once    sync.Once
}

func New(sources ...types.EventSource) *ControlSource {
	return &ControlSource{
		events:  make(chan types.Event, 128),
		sources: sources,
	}
}

func (c *ControlSource) startSources(ctx context.Context) {
	c.once.Do(func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for _, src := range c.sources {
			c.forwardSource(ctx, src)
		}
	})
}

func (c *ControlSource) forwardSource(ctx context.Context, src types.EventSource) {
	go func() {
		err := src.Serve(ctx, c.events)
		// NOTE: runs only when the source stops serving
		if err != nil && ctx.Err() == nil {
			select {
			case c.events <- types.ErrorEvent{Err: err}:
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *ControlSource) Serve(ctx context.Context, out chan<- types.Event) error {
	c.startSources(ctx)

	for {
		// NOTE: check cancellation before waiting
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// NOTE: wait for either event or cancellation
		select {
		case ev := <-c.events:
			// NOTE: check cancellation while waiting for write to out
			select {
			case out <- ev:

			case <-ctx.Done():
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}
