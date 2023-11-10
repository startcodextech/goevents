package eventsource

import (
	"context"
	"github.com/start-codex/goevents/ddd"
	"github.com/start-codex/goutils/id"
)

type (
	EventSourcedAggregate interface {
		id.IDer
		AggregateName() string
		ddd.Eventer
		Versioner
		EventApplier
		EventCommitter
	}

	AggregateStoreMiddleware func(AggregateStore) AggregateStore

	AggregateStore interface {
		Load(context.Context, EventSourcedAggregate) error
		Save(context.Context, EventSourcedAggregate) error
	}
)

func AggregateStoreWithMiddleware(store AggregateStore, middleware ...AggregateStoreMiddleware) AggregateStore {
	s := store
	// Middlewares are applied in a reverse sequence; this positions the first middleware
	// in the array as the outermost layer, meaning it's the first to be entered and the last to exit.
	// Given: store, A, B, C
	// Outcome: A(B(C(store))
	for i := len(middleware) - 1; i >= 0; i-- {
		s = middleware[i](s)
	}
	return s
}
