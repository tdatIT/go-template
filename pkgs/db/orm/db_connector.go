package orm

import (
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/tdatIT/go-template/config"
	"go.uber.org/zap"
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
		logger.Fatal("error while creating db connection", zap.Error(err))
	}

	logger.Info("db connection established",
		zap.String("host", fmt.Sprintf("%v:%v", config.Database.Host, config.Database.Port)),
		zap.String("db", config.Database.Database),
		zap.String("schema", config.Database.Schema),
	)

	return conn
}
