package sql

import (
	"context"
	"fmt"
	"github.com/startcodextech/goevents/registry"
	"github.com/startcodextech/goevents/sec"
)

type SagaStore struct {
	tableName string
	db        DB
	registry  registry.Registry
}

var _ sec.SagaStore = (*SagaStore)(nil)

func NewSagaStore(tableName string, db DB, registry registry.Registry) SagaStore {
	return SagaStore{
		tableName: tableName,
		db:        db,
		registry:  registry,
	}
}

func (s SagaStore) Load(ctx context.Context, sagaName, sagaID string) (*sec.SagaContext[[]byte], error) {
	sagaCtx := &sec.SagaContext[[]byte]{
		ID: sagaID,
	}
	err := s.db.QueryRowContext(ctx, s.table(s.queryLoad()), sagaName, sagaID).Scan(&sagaCtx.Data, &sagaCtx.Step, &sagaCtx.Done, &sagaCtx.Compensating)

	return sagaCtx, err
}

func (s SagaStore) Save(ctx context.Context, sagaName string, sagaCtx *sec.SagaContext[[]byte]) error {
	_, err := s.db.ExecContext(ctx, s.table(s.querySave()), sagaName, sagaCtx.ID, sagaCtx.Data, sagaCtx.Step, sagaCtx.Done, sagaCtx.Compensating)

	return err
}

func (s SagaStore) queryLoad() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "SELECT data, step, done, compensating FROM %s WHERE name = ? AND id = ?"
	default:
		return "SELECT data, step, done, compensating FROM %s WHERE name = $1 AND id = $2"
	}
}

func (s SagaStore) querySave() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return `INSERT INTO %s (name, id, data, step, done, compensating) 
					VALUES (?, ?, ?, ?, ?, ?) 
				ON DUPLICATE KEY UPDATE
					data = VALUES(data), step = VALUES(step), done = VALUES(done), compensating = VALUES(compensating)`
	default:
		return `INSERT INTO %s (name, id, data, step, done, compensating) 
					VALUES ($1, $2, $3, $4, $5, $6) 
				ON CONFLICT (name, id) DO
					UPDATE SET data = EXCLUDED.data, step = EXCLUDED.step, done = EXCLUDED.done, compensating = EXCLUDED.compensating`
	}
}

func (s SagaStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
