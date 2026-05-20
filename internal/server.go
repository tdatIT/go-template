package server

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/tdatIT/go-template/config"
	userApp "github.com/tdatIT/go-template/internal/app/user"
	userHandler "github.com/tdatIT/go-template/internal/handler/user"
	userRepos "github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/internal/router"
	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
	"github.com/tdatIT/go-template/pkgs/logger"
)

type Server struct {
	cfg         *config.AppConfig
	echo        *echo.Echo
	database    orm.ORM
	redisClient rdclient.RedisClient
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

	database := orm.NewDBConnection(cfg)
	redisClient := rdclient.NewRedisClient(cfg)

	echoApp := newHttpServer(cfg)

	userRepository := userRepos.NewUserRepository(database)

	userApplication := userApp.NewUserApplication(userRepository)

	usrHandle := userHandler.NewUserHandler(userApplication)

	router.RegisterRoutes(echoApp, usrHandle)

	return &Server{
		cfg:         cfg,
		echo:        echoApp,
		database:    database,
		redisClient: redisClient,
	}
}

func (serv Server) API() *echo.Echo {
	return serv.echo
}

func (serv Server) Config() *config.AppConfig {
	return serv.cfg
}

func (serv Server) Shutdown() {
	if serv.database != nil {
		if err := serv.database.Close(); err != nil {
			slog.Error("failed to close database", slog.String("error", err.Error()))
		}
	}

	if serv.redisClient != nil {
		if err := serv.redisClient.Close(); err != nil {
			slog.Error("failed to close redis client", slog.String("error", err.Error()))
		}
	}
}
