package orm

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/tdatIT/go-template/config"
)

func NewDBConnection(config *config.AppConfig) ORM {
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		config.Database.Host,
		config.Database.Port,
		config.Database.UserName,
		config.Database.Password,
		config.Database.Database,
		config.Database.Options,
	)

	if config.Database.Schema != "" {
		dataSourceName += fmt.Sprintf(" search_path=%s", config.Database.Schema)
	}

	cfg := Config{
		DSN:             dataSourceName,
		MaxOpenConns:    config.Database.MaxOpenConns,
		MaxIdleConns:    config.Database.MaxIdleConns,
		ConnMaxLifetime: config.Database.ConnMaxLifetime,
		ConnMaxIdleTime: config.Database.ConnMaxIdleTime,
		Debug:           config.Database.Debug,
	}

	conn, err := newGormInstance(cfg)
	if err != nil {
		slog.Error("error while creating db connection", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("db connection established",
		slog.String("host", fmt.Sprintf("%v:%v", config.Database.Host, config.Database.Port)),
		slog.String("db", config.Database.Database),
		slog.String("schema", config.Database.Schema),
	)

	return conn
}
