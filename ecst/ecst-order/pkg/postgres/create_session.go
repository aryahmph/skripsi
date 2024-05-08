package postgres

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"net/url"
)

func CreateSession(cfg *Config) (*sqlx.DB, error) {
	param := url.Values{}
	param.Add("sslmode", "disable")

	connStr := fmt.Sprintf(connStringTemplate,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		param.Encode(),
	)

	db, err := sqlx.Open("pgx", connStr)
	if err != nil {
		return db, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return db, nil
}
