package server

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/tdatIT/go-template/config"
	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
	"github.com/tdatIT/go-template/pkgs/logger"
)

type Server struct {
	cfg      *config.AppConfig
	echo     *echo.Echo
	database orm.ORM
	redis    rdclient.RedisClient
}

func NewServer() *Server {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.SetDefault(slog.New(logger.NewJsonSlogHandler(
		&logger.SlogConfig{
			Level:       cfg.Logger.Level,
			ServiceName: cfg.Server.Name,
		},
		true,
	)))

	//int echo
	echoApp := newHttpServer(cfg)
	////db connection
	//database := orm.NewDBConnection(cfg)
	//redisDb := rdclient.NewRedisClient(cfg)

	return &Server{
		cfg:  cfg,
		echo: echoApp,
		//_database: database,
		//_redis:    redisDb,
	}
}

func (serv Server) API() *echo.Echo {
	return serv.echo
}

func (serv Server) Config() *config.AppConfig {
	return serv.cfg
}

func (serv Server) Shutdown() {
	if err := serv.database.Close(); err != nil {
		slog.Error("failed to close database", slog.String("error", err.Error()))
	}

	if err := serv.redis.Close(); err != nil {
		slog.Error("failed to close redis client", slog.String("error", err.Error()))
	}
}
