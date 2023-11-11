package mongo

import (
	"context"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/eventsourcing"
	"github.com/startcodextech/goevents/registry"
	"github.com/startcodextech/goevents/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	EventStore struct {
		collection Collection
		registry   registry.Registry
	}
)

var _ eventsourcing.AggregateStore = (*EventStore)(nil)

func NewEventStore(collectionName string, collection Collection, registry registry.Registry) EventStore {
	return EventStore{
		collection: collection,
		registry:   registry,
	}
}

func (s EventStore) Load(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) (err error) {
	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()

	filter := bson.D{
		{"stream_id", aggregateID},
		{"stream_name", aggregateName},
		{"stream_version", bson.D{{"$gt", aggregate.Version()}}},
	}

	findOptions := options.Find().SetSort(bson.D{{"stream_version", 1}})

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return err
	}

	defer func(cursor *mongo.Cursor) {
		err := cursor.Close(ctx)
		if err != nil {
			err = errors.Wrap(err, "closing event cursor")
		}
	}(cursor)

	for cursor.Next(ctx) {
		var event store.AggregateEvent
		if err = cursor.Decode(&event); err != nil {
			return err
		}

		if err = eventsourcing.LoadEvent(aggregate, event); err != nil {
			return err
		}
	}
	return nil
}

func (s EventStore) Save(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) (err error) {
	aggregateID := aggregate.ID()
	aggregateName := aggregate.AggregateName()

	documents := make([]interface{}, len(aggregate.Events()))

	for i, event := range aggregate.Events() {
		var payloadData []byte
		payloadData, err = s.registry.Serialize(event.EventName(), event.Payload())
		if err != nil {
			return err
		}

		documents[i] = bson.M{
			"stream_id":      aggregateID,
			"stream_name":    aggregateName,
			"stream_version": event.AggregateVersion(),
			"event_id":       event.ID(),
			"event_name":     event.EventName(),
			"event_data":     payloadData,
			"occurred_at":    event.OccurredAt(),
		}
	}

	_, err = s.collection.InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	return nil
}
