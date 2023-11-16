package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/async"
	"github.com/startcodextech/goevents/transmanager"
)

type (
	InboxStore struct {
		tableName string
		db        DB
	}
)

var _ transmanager.InboxStore = (*InboxStore)(nil)

func NewInboxStore(tableName string, db DB) InboxStore {
	return InboxStore{
		tableName: tableName,
		db:        db,
	}
}

func (s InboxStore) Save(ctx context.Context, msg async.IncomingMessage) error {
	metadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, s.table(s.querySave()), msg.ID(), msg.MessageName(), msg.Subject(), msg.Data(), metadata, msg.SentAt(), msg.ReceivedAt())
	if err != nil {
		switch s.db.DBType() {
		case DBTypeMySQL:
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return transmanager.ErrDuplicateMessage(msg.ID())
				}
			}
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					return transmanager.ErrDuplicateMessage(msg.ID())
				}
			}
		}

	}

	return err
}

func (s InboxStore) querySave() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "INSERT INTO %s (id, NAME, subject, DATA, metadata, sent_at, received_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	default:
		return "INSERT INTO %s (id, NAME, subject, DATA, metadata, sent_at, received_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	}
}

func (s InboxStore) table(query string) string {
	return fmt.Sprintf(query, s.tableName)
}
