package sqlotel

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/stackus/errors"
	sql2 "github.com/startcodextech/goevents/store/sql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"time"
)

type tracedDB struct {
	db sql2.DB
}

var _ sql2.DB = (*tracedDB)(nil)

func Trace(db sql2.DB) sql2.DB {
	return tracedDB{db: db}
}

func (t tracedDB) PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("PrepareContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, err)
	}(time.Now())

	return t.db.PrepareContext(ctx, query)
}

func (t tracedDB) ExecContext(ctx context.Context, query string, args ...any) (result sql.Result, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("ExecContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, err)
	}(time.Now())

	return t.db.ExecContext(ctx, query, args...)
}

func (t tracedDB) QueryContext(ctx context.Context, query string, args ...any) (rows *sql.Rows, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("QueryContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, err)
	}(time.Now())

	return t.db.QueryContext(ctx, query, args...)
}

func (t tracedDB) QueryRowContext(ctx context.Context, query string, args ...any) (row *sql.Row) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("QueryRowContext", trace.WithAttributes(
			attribute.String("Query", query),
			attribute.Float64("Took", time.Since(started).Seconds()),
		))
		t.recordError(span, row.Err())
	}(time.Now())

	return t.db.QueryRowContext(ctx, query, args...)
}

func (t tracedDB) recordError(span trace.Span, err error) {
	if err != nil {
		isError := false
		switch t.DBType() {
		case sql2.DBTypeMySQL:
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				isError = true
				span.AddEvent("Database Error", trace.WithAttributes(
					attribute.String("Error", err.Error()),
					attribute.String("Code", strconv.Itoa(int(mysqlErr.Number))),
					attribute.String("Message", mysqlErr.Message),
				))
			}
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				isError = true
				span.AddEvent("Database Error", trace.WithAttributes(
					attribute.String("Error", err.Error()),
					attribute.String("Code", pgErr.Code),
					attribute.String("Severity", pgErr.Severity),
					attribute.String("Message", pgErr.Message),
					attribute.String("Detail", pgErr.Detail),
				))
			}
		}

		if !isError {
			span.AddEvent("Database Error", trace.WithAttributes(
				attribute.String("Error", err.Error()),
			))
		}
	}
}

func (t tracedDB) DBType() string {
	return t.db.DBType()
}
