package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/eventsourcing"
	"github.com/startcodextech/goevents/registry"
	"github.com/startcodextech/goevents/store"
	"strings"
	"time"
)

type (
	EventStore struct {
		tableName string
		db        DB
		registry  registry.Registry
	}
)

var _ eventsourcing.AggregateStore = (*EventStore)(nil)

func NewEventStore(tableName string, db DB, registry registry.Registry) EventStore {
	return EventStore{
		tableName: tableName,
		db:        db,
		registry:  registry,
	}
}

func (s EventStore) Load(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) (err error) {
	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()

	var rows *sql.Rows

	rows, err = s.db.QueryContext(ctx, s.table(s.queryLoad()), aggregateID, aggregateName, aggregate.Version())
	if err != nil {
		return err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			err = errors.Wrap(err, "closing event rows")
		}
	}(rows)

	for rows.Next() {
		var eventID, eventName string
		var payloadData []byte
		var aggregateVersion int
		var occurredAt time.Time
		err := rows.Scan(&aggregateVersion, &eventID, &eventName, &payloadData, &occurredAt)
		if err != nil {
			return err
		}

		var payload interface{}
		payload, err = s.registry.Deserialize(eventName, payloadData)
		if err != nil {
			return err
		}

		event := store.NewAggregateEventBuilder().
			WithID(eventID).
			WithName(eventName).
			WithPayload(payload).
			WithAggregate(aggregate).
			WithAggregateVersion(aggregateVersion).
			WithOccurredAt(occurredAt).
			Build()

		if err = eventsourcing.LoadEvent(aggregate, event); err != nil {
			return err
		}
	}
	return nil
}

func (s EventStore) Save(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) (err error) {
	const query = "INSERT INTO %s (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at) VALUES"

	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()

	placeholders := make([]string, len(aggregate.Events()))
	values := make([]any, len(aggregate.Events())*7)

	for i, event := range aggregate.Events() {
		var payloadData []byte

		payloadData, err := s.registry.Serialize(event.EventName(), event.Payload())
		if err != nil {
			return err
		}

		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7,
		)

		values[i*7] = aggregateID
		values[i*7+1] = aggregateName
		values[i*7+2] = event.AggregateVersion()
		values[i*7+3] = event.ID()
		values[i*7+4] = event.EventName()
		values[i*7+5] = payloadData
		values[i*7+6] = event.OccurredAt()
	}

	_, err = s.db.ExecContext(
		ctx,
		fmt.Sprintf("%s %s", s.table(query), strings.Join(placeholders, ",")),
		values...,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s EventStore) queryLoad() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "SELECT stream_version, event_id, event_name, event_data, occurred_a FROM %s WHERE stream_id = ? AND stream_name = ? AND stream_version > ? ORDER BY stream_version AS"
	default:
		return "SELECT stream_version, event_id, event_name, event_data, occurred_a FROM %s WHERE stream_id = $1 AND stream_name = $2 AND stream_version > $3 ORDER BY stream_version ASC"

	}
}

func (s EventStore) querySave() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "INSERT INTO %s (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at) VALUES"
	default:
		return "INSERT INTO %s (stream_id, stream_name, stream_version, event_id, event_name, event_data, occurred_at) VALUES"

	}
}

func (s EventStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
