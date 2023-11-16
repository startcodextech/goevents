package mongo

import (
	"context"
	"fmt"
	"github.com/startcodextech/goevents/esourcing"
	"github.com/startcodextech/goevents/registry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SnapshotStore struct {
	esourcing.AggregateStore
	collection Collection
	registry   registry.Registry
}

var _ esourcing.AggregateStore = (*SnapshotStore)(nil)

func NewSnapshotStore(collection Collection, registry registry.Registry) esourcing.AggregateStoreMiddleware {
	snapshots := SnapshotStore{
		collection: collection,
		registry:   registry,
	}

	return func(store esourcing.AggregateStore) esourcing.AggregateStore {
		snapshots.AggregateStore = store
		return snapshots
	}
}

func (s SnapshotStore) Load(ctx context.Context, aggregate esourcing.EventSourcedAggregate) error {
	filter := bson.M{"stream_id": aggregate.ID(), "stream_name": aggregate.AggregateName()}
	opts := options.FindOne().SetSort(bson.D{{"stream_version", -1}}) // To get the latest snapshot

	var result struct {
		StreamVersion int    `bson:"stream_version"`
		SnapshotName  string `bson:"snapshot_name"`
		SnapshotData  []byte `bson:"snapshot_data"`
	}

	err := s.collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Your logic for when no snapshot is found
		}
		return err
	}

	v, err := s.registry.Deserialize(result.SnapshotName, result.SnapshotData, registry.ValidateImplements((*esourcing.Snapshot)(nil)))
	if err != nil {
		return err
	}

	if err := esourcing.LoadSnapshot(aggregate, v.(esourcing.Snapshot), result.StreamVersion); err != nil {
		return err
	}

	return s.AggregateStore.Load(ctx, aggregate)
}

func (s SnapshotStore) Save(ctx context.Context, aggregate esourcing.EventSourcedAggregate) error {
	if err := s.AggregateStore.Save(ctx, aggregate); err != nil {
		return err
	}

	if !s.shouldSnapshot(aggregate) {
		return nil
	}

	sser, ok := aggregate.(esourcing.Snapshotter)
	if !ok {
		return fmt.Errorf("%T does not implement esourcing.Snapshotter", aggregate)
	}

	snapshot := sser.ToSnapshot()

	data, err := s.registry.Serialize(snapshot.SnapshotName(), snapshot)
	if err != nil {
		return err
	}

	_, err = s.collection.UpdateOne(
		ctx,
		bson.M{"stream_id": aggregate.ID(), "stream_name": aggregate.AggregateName()},
		bson.M{
			"$set": bson.M{
				"stream_version": aggregate.PendingVersion(),
				"snapshot_name":  snapshot.SnapshotName(),
				"snapshot_data":  data,
			},
		},
		options.Update().SetUpsert(true),
	)

	return err
}

// TODO use injected & configurable strategies
func (SnapshotStore) shouldSnapshot(aggregate esourcing.EventSourcedAggregate) bool {
	var maxChanges = 3 // low for demonstration; production envs should use higher values 50, 75, 100...
	var pendingVersion = aggregate.PendingVersion()
	var pendingChanges = len(aggregate.Events())

	return pendingVersion >= maxChanges && ((pendingChanges >= maxChanges) ||
		(pendingVersion%maxChanges < pendingChanges) ||
		(pendingVersion%maxChanges == 0))
}
