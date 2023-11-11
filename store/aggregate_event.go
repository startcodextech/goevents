package store

import (
	"github.com/startcodextech/goevents/ddd"
	"github.com/startcodextech/goevents/eventsourcing"
	"time"
)

type (
	AggregateEvent struct {
		id         string
		name       string
		payload    ddd.EventPayload
		occurredAt time.Time
		aggregate  eventsourcing.EventSourcedAggregate
		version    int
	}
)

var _ ddd.AggregateEvent = (*AggregateEvent)(nil)

func NewAggregateEventBuilder() *AggregateEvent {
	return &AggregateEvent{}
}

func (e *AggregateEvent) Build() AggregateEvent {
	return *e
}

func (e *AggregateEvent) WithID(id string) *AggregateEvent {
	e.id = id
	return e
}

func (e *AggregateEvent) WithName(name string) *AggregateEvent {
	e.name = name
	return e
}

func (e *AggregateEvent) WithPayload(payload ddd.EventPayload) *AggregateEvent {
	e.payload = payload
	return e
}

func (e *AggregateEvent) WithAggregate(aggregate eventsourcing.EventSourcedAggregate) *AggregateEvent {
	e.aggregate = aggregate
	return e
}

func (e *AggregateEvent) WithAggregateVersion(version int) *AggregateEvent {
	e.version = version
	return e
}

func (e *AggregateEvent) WithOccurredAt(occurredAt time.Time) *AggregateEvent {
	e.occurredAt = occurredAt
	return e
}

func (e AggregateEvent) ID() string                { return e.id }
func (e AggregateEvent) EventName() string         { return e.name }
func (e AggregateEvent) Payload() ddd.EventPayload { return e.payload }
func (e AggregateEvent) Metadata() ddd.Metadata    { return ddd.Metadata{} }
func (e AggregateEvent) OccurredAt() time.Time     { return e.occurredAt }
func (e AggregateEvent) AggregateName() string     { return e.aggregate.AggregateName() }
func (e AggregateEvent) AggregateID() string       { return e.aggregate.ID() }
func (e AggregateEvent) AggregateVersion() int     { return e.version }
