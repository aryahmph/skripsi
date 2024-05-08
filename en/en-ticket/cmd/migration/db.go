package migration

import (
	"en-ticket/internal/appctx"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/postgres"
	"time"
)

func MigrateDatabase() {
	cfg, e := appctx.NewConfig()
	if e != nil {
		logger.Fatal(e)
	}

	postgres.DatabaseMigration(&postgres.Config{
		Host:         cfg.DBWrite.Host,
		Port:         cfg.DBWrite.Port,
		Name:         cfg.DBWrite.Name,
		User:         cfg.DBWrite.User,
		Password:     cfg.DBWrite.Pass,
		Charset:      cfg.DBWrite.Charset,
		Timeout:      time.Duration(cfg.DBWrite.TimeoutSecond) * time.Second,
		MaxIdleConns: cfg.DBWrite.MaxIdle,
		MaxOpenConns: cfg.DBWrite.MaxOpen,
		MaxLifetime:  time.Duration(cfg.DBWrite.MaxLifeTimeMS) * time.Millisecond,
	})
}
