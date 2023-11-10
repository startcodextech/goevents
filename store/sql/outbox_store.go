package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/stackus/errors"
	"github.com/start-codex/goevents/asyncmessages"
	"github.com/start-codex/goevents/store"
	"github.com/start-codex/goevents/transactionmanager"
	"time"
)

type (
	OutboxStore struct {
		tableName string
		db        DB
	}
)

var _ transactionmanager.OutboxStore = (*OutboxStore)(nil)

func NewOutboxStore(tableName string, db DB) OutboxStore {
	return OutboxStore{
		tableName: tableName,
		db:        db,
	}
}

func (s OutboxStore) Save(ctx context.Context, msg asyncmessages.Message) error {

	metadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, s.table(s.querySave()), msg.ID(), msg.MessageName(), msg.Subject(), msg.Data(), metadata, msg.SentAt())
	if err != nil {
		switch s.db.DBType() {
		case DBTypeMySQL:
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return transactionmanager.ErrDuplicateMessage(msg.ID())
				}
			}
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					return transactionmanager.ErrDuplicateMessage(msg.ID())
				}
			}
		}
	}

	return err
}

func (s OutboxStore) FindUnpublished(ctx context.Context, limit int) ([]asyncmessages.Message, error) {
	rows, err := s.db.QueryContext(ctx, s.table(s.queryFindUnpublished(), limit))
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			err = errors.Wrap(err, "closing event rows")
		}
	}(rows)

	var msgs []asyncmessages.Message

	for rows.Next() {
		var metadata []byte
		var id, name, subject string
		var data []byte
		var sentAt time.Time
		err = rows.Scan(&id, &name, &subject, &data, &metadata, &sentAt)
		if err != nil {
			return msgs, err
		}

		var mdata map[string]any

		err = json.Unmarshal(metadata, &mdata)

		msg := store.NewOutboxMessageBuilder().
			WithID(id).
			WithName(name).
			WithSubject(subject).
			WithData(data).
			WithMetadata(mdata).
			WithSendAt(sentAt).
			Build()

		msgs = append(msgs, msg)
	}

	return msgs, rows.Err()
}

func (s OutboxStore) MarkPublished(ctx context.Context, ids ...string) (err error) {

	switch s.db.DBType() {
	case DBTypeMySQL:
		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}
		_, err = s.db.ExecContext(ctx, s.table(s.queryMarkPublished()), args...)
	default:
		msgIDs := &pgtype.TextArray{}
		err := msgIDs.Set(ids)
		if err != nil {
			return err
		}

		_, err = s.db.ExecContext(ctx, s.table(s.queryMarkPublished()), msgIDs)
	}

	return err
}

func (s OutboxStore) querySave() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "INSERT INTO %s (id, NAME, subject, DATA, metadata, sent_at) VALUES (?, ?, ?, ?, ?, ?)"
	default:
		return "INSERT INTO %s (id, NAME, subject, DATA, metadata, sent_at) VALUES ($1, $2, $3, $4, $5, $6)"
	}
}

func (s OutboxStore) queryFindUnpublished() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "SELECT id, name, subject, data, metadata, sent_at FROM %s WHERE published_at IS NULL LIMIT %d"
	default:
		return "SELECT id, name, subject, data, metadata, sent_at FROM %s WHERE published_at IS NULL LIMIT %d"
	}
}

func (s OutboxStore) queryMarkPublished() string {
	switch s.db.DBType() {
	case DBTypeMySQL:
		return "UPDATE %s SET published_at = CURRENT_TIMESTAMP WHERE id IN (?)"
	default:
		return "UPDATE %s SET published_at = CURRENT_TIMESTAMP WHERE id = ANY ($1)"
	}
}

func (s OutboxStore) table(query string, args ...any) string {
	params := []any{s.tableName}
	params = append(params, args...)
	return fmt.Sprintf(query, params...)
}
