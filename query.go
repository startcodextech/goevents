package goevents

import (
	"context"
)

type (
	Payload []byte

	Query struct {
		Command
	}

	QueryHandler          func(ctx context.Context, query Query) (Payload, error)
	QuerySubscribeHandler func(ctx context.Context, query Query, payload Payload)

	QueryBus interface {
		Subscribe(query string, handler QuerySubscribeHandler)
		Publish(ctx context.Context, query Query) (Payload, error)
		RegisterHandler(query string, handler QueryHandler)
	}
)

func CreateQuery(name string, params Params) Query {
	return Query{
		Command: CreateCommand(name, params),
	}
}

func CreateQueryCorrelationID(uuid, name string, params Params) Query {
	return Query{
		Command: CreateCommandCorrelationID(uuid, name, params),
	}
}
