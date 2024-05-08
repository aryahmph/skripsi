package bootstrap

import (
	"time"

	config "ecst-ticket/internal/appctx"
	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/postgres"
)

func RegistryPostgresDBSingle(cfg *config.Database, timezone string) postgres.Adapter {
	db, err := postgres.NewPostgresSingle(
		&postgres.Config{
			Host:         cfg.Host,
			Name:         cfg.Name,
			Password:     cfg.Pass,
			Port:         cfg.Port,
			User:         cfg.User,
			Timeout:      time.Duration(cfg.TimeoutSecond) * time.Second,
			MaxOpenConns: cfg.MaxOpen,
			MaxIdleConns: cfg.MaxIdle,
			MaxLifetime:  time.Duration(cfg.MaxLifeTimeMS) * time.Millisecond,
			Charset:      cfg.Charset,
			TimeZone:     timezone,
		},
	)

	if err != nil {
		logger.Fatal(err,
			logger.EventName("db"),
			logger.Any("host_write", cfg.Host),
			logger.Any("port_write", cfg.Port),
		)
	}

	return db
}
