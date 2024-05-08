package postgres

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type postgresSingle struct {
	db  *sqlx.DB
	cfg *Config
}

func NewPostgresSingle(cfg *Config) (Adapter, error) {
	x := postgresSingle{cfg: cfg}

	e := x.initialize()

	return &x, e
}

func (d *postgresSingle) initialize() error {
	dbWrite, err := CreateSession(d.cfg)

	if err != nil {
		return err
	}

	err = dbWrite.Ping()
	if err != nil {
		return err
	}

	d.db = dbWrite

	return nil
}

// QueryRow select single row database will return  sql.row raw
func (d *postgresSingle) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// QueryRows select multiple rows of database will return  sql.rows raw
func (d *postgresSingle) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// Fetch select multiple rows of database will cast data to struct passing by parameter
func (d *postgresSingle) Fetch(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return d.db.SelectContext(ctx, dst, query, args...)
}

// FetchRow fetching one row database will cast data to struct passing by parameter
func (d *postgresSingle) FetchRow(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return d.db.GetContext(ctx, dst, query, args...)
}

// Exec execute mysql command query
func (d *postgresSingle) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// BeginTx start new transaction session
func (d *postgresSingle) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

// Ping check database connectivity
func (d *postgresSingle) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// HealthCheck checking healthy of database connection
func (d *postgresSingle) HealthCheck() error {
	return d.Ping(context.Background())
}
