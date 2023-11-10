package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/eventsourcing"
	"github.com/startcodextech/goevents/registry"
)

type SnapshotStore struct {
	eventsourcing.AggregateStore
	tableName string
	db        DB
	registry  registry.Registry
}

var _ eventsourcing.AggregateStore = (*SnapshotStore)(nil)

func NewSnapshotStore(tableName string, db DB, registry registry.Registry) eventsourcing.AggregateStoreMiddleware {
	snapshots := SnapshotStore{
		tableName: tableName,
		db:        db,
		registry:  registry,
	}

	return func(store eventsourcing.AggregateStore) eventsourcing.AggregateStore {
		snapshots.AggregateStore = store
		return snapshots
	}
}

func (s SnapshotStore) Load(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) error {

	var entityVersion int
	var snapshotName string
	var snapshotData []byte

	if err := s.db.QueryRowContext(ctx, s.table(s.queryLoad()), aggregate.ID(), aggregate.AggregateName()).Scan(&entityVersion, &snapshotName, &snapshotData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.AggregateStore.Load(ctx, aggregate)
		}
		return err
	}

	v, err := s.registry.Deserialize(snapshotName, snapshotData, registry.ValidateImplements((*eventsourcing.Snapshot)(nil)))
	if err != nil {
		return err
	}

	if err := eventsourcing.LoadSnapshot(aggregate, v.(eventsourcing.Snapshot), entityVersion); err != nil {
		return err
	}

	return s.AggregateStore.Load(ctx, aggregate)
}

func (s SnapshotStore) Save(ctx context.Context, aggregate eventsourcing.EventSourcedAggregate) error {
	if err := s.AggregateStore.Save(ctx, aggregate); err != nil {
		return err
	}

	if !s.shouldSnapshot(aggregate) {
		return nil
	}

	sser, ok := aggregate.(eventsourcing.Snapshotter)
	if !ok {
		return fmt.Errorf("%T does not implelement eventsourcing.Snapshotter", aggregate)
	}

	snapshot := sser.ToSnapshot()

	data, err := s.registry.Serialize(snapshot.SnapshotName(), snapshot)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, s.table(s.querySave()), aggregate.ID(), aggregate.AggregateName(), aggregate.PendingVersion(), snapshot.SnapshotName(), data)

	return err
}

// TODO use injected & configurable strategies
func (SnapshotStore) shouldSnapshot(aggregate eventsourcing.EventSourcedAggregate) bool {
	var maxChanges = 3 // low for demonstration; production envs should use higher values 50, 75, 100...
	var pendingVersion = aggregate.PendingVersion()
	var pendingChanges = len(aggregate.Events())

	return pendingVersion >= maxChanges && ((pendingChanges >= maxChanges) ||
		(pendingVersion%maxChanges < pendingChanges) ||
		(pendingVersion%maxChanges == 0))
}

func (s SnapshotStore) queryLoad() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "SELECT stream_version, snapshot_name, snapshot_data FROM %s WHERE stream_id = ? AND stream_name = ? LIMIT 1"
	default:
		return "SELECT stream_version, snapshot_name, snapshot_data FROM %s WHERE stream_id = $1 AND stream_name = $2 LIMIT 1"
	}
}

func (s SnapshotStore) querySave() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return `INSERT INTO %s (stream_id, stream_name, stream_version, snapshot_name, snapshot_data) 
					VALUES (?, ?, ?, ?, ?) 
				ON DUPLICATE KEY UPDATE 
					stream_version = VALUES(stream_version), snapshot_name = VALUES(snapshot_name), snapshot_data = VALUES(snapshot_data)
`
	default:
		return `INSERT INTO %s (stream_id, stream_name, stream_version, snapshot_name, snapshot_data) 
					VALUES ($1, $2, $3, $4, $5) 
				ON CONFLICT (stream_id, stream_name) DO
					UPDATE SET stream_version = EXCLUDED.stream_version, snapshot_name = EXCLUDED.snapshot_name, snapshot_data = EXCLUDED.snapshot_data`
	}
}

func (s SnapshotStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
