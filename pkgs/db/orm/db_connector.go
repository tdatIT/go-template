package orm

import (
	"fmt"
	"log/slog"

	"github.com/tdatIT/go-template/config"
	"github.com/tdatIT/go-template/pkgs/logger"
)

func NewDBConnection(config *config.AppConfig) (ORM, error) {
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
		Logger: logger.NewSlogLogger(&logger.SlogConfig{
			Level:       config.Logger.Level,
			ServiceName: config.Server.Name,
			Format:      config.Logger.Format,
		}, false),
	}

	conn, err := newGormInstance(cfg)
	if err != nil {
		slog.Error("error while creating db connection", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("db connection created",
		slog.String("host", fmt.Sprintf("%v:%v", config.Database.Host, config.Database.Port)),
		slog.String("db", config.Database.Database),
		slog.String("schema", config.Database.Schema),
	)

	return conn, nil
}
