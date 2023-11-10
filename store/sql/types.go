package sql

import (
	"context"
	"database/sql"
)

const (
	DBTypePostgres = "PostgreSQL"
	DBTypeMySQL    = "MySQL"
)

type (
	DB interface {
		PrepareContext(context.Context, string) (*sql.Stmt, error)
		ExecContext(context.Context, string, ...any) (sql.Result, error)
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
		QueryRowContext(context.Context, string, ...any) *sql.Row
		DBType() string
	}

	PostgresDB struct {
		sql.DB
	}

	MySQLDB struct {
		sql.DB
	}
)

var _ (DB) = (*PostgresDB)(nil)

var _ (DB) = (*MySQLDB)(nil)

func (p *PostgresDB) DBType() string {
	return DBTypePostgres
}

func (m *MySQLDB) DBType() string {
	return DBTypeMySQL
}
