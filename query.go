package goevents

import (
	"context"
)

type (
	Payload interface{}

	Query struct {
		Command
	}

	QueryHandler          func(ctx context.Context, query Query) (Payload, error)
	QuerySubscribeHandler func(ctx context.Context, query Query, payload Payload)

	QueryBus interface {
		Subscribe(query string, handler QuerySubscribeHandler)
		Publish(ctx context.Context, query Query) error
		RegisterHandler(query string, handler QueryHandler)
	}
)

func CreateQuery(name string, payload Payload) Query {
	return Query{
		Command: CreateCommand(name, payload),
	}
}

func CreateQueryCorrelationID(uuid, name string, params Payload) Query {
	return Query{
		Command: CreateCommandCorrelationID(uuid, name, params),
	}
}
