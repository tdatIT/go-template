package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"

	"github.com/tdatIT/go-template/internal/worker/event"

	"github.com/tdatIT/go-template/config"
	userApp "github.com/tdatIT/go-template/internal/app/user"
	userHandler "github.com/tdatIT/go-template/internal/handler/user"
	"github.com/tdatIT/go-template/internal/infras/adapter/productsvc"
	"github.com/tdatIT/go-template/internal/infras/mqttpub"
	userRepos "github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/internal/router"
	"github.com/tdatIT/go-template/internal/worker"
	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
	"github.com/tdatIT/go-template/pkgs/logger"
	mqttpkg "github.com/tdatIT/go-template/pkgs/mqtt"
	"github.com/tdatIT/go-template/pkgs/tracing"
)

type Server struct {
	cfg         *config.AppConfig
	echo        *echo.Echo
	database    orm.ORM
	redisClient rdclient.RedisClient
	mqttClient  mqttpkg.Client
	workerGroup *worker.WorkerGroup
	tracerStop  func(context.Context) error
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

	// Tracing must be initialized before the HTTP server so the middleware
	// (registered in newHttpServer) picks up the global TracerProvider/propagator.
	var tracerStop func(context.Context) error
	if cfg.Tracing.Enabled {
		stop, traceErr := tracing.Init(context.Background(), tracing.Config{
			Endpoint:       cfg.Tracing.Endpoint,
			Insecure:       cfg.Tracing.Insecure,
			SampleRatio:    cfg.Tracing.SampleRatio,
			ServiceName:    cfg.Server.Name,
			ServiceVersion: cfg.Server.Version,
		})
		if traceErr != nil {
			slog.Error("failed to init tracing", slog.String("error", traceErr.Error()))
			os.Exit(1)
		}
		tracerStop = stop
	}

	database := orm.NewDBConnection(cfg)
	redisClient := rdclient.NewRedisClient(cfg)

	echoApp := newHttpServer(cfg)

	userRepository := userRepos.NewUserRepository(database)
	productAdapter := productsvc.NewAdapter(&cfg.Adapters.ProductService)
	userApplication := userApp.NewUserApplication(userRepository, productAdapter)

	usrHandle := userHandler.NewUserHandler(userApplication)
	router.RegisterRoutes(echoApp, usrHandle)

	// WorkerGroup is always created; workers are only registered when MQTT is available.
	workerGroup := worker.NewWorkerGroup()

	mqttCli, conErr := mqttpkg.NewMQTTClient(cfg)
	if conErr != nil {
		slog.Warn("mqtt client unavailable, workers will not start", slog.String("error", conErr.Error()))
	} else {
		// Publisher is available for injection into the app layer when needed.
		_ = mqttpub.NewPublisher(mqttCli)

		// Register all topic workers here.
		workerGroup.Register(
			event.NewUserEventWorker(mqttCli, userApplication),
		)
	}

	return &Server{
		cfg:         cfg,
		echo:        echoApp,
		database:    database,
		redisClient: redisClient,
		mqttClient:  mqttCli,
		workerGroup: workerGroup,
		tracerStop:  tracerStop,
	}
}

// API returns the Echo HTTP server component.
func (s *Server) API() *echo.Echo {
	return s.echo
}

// Workers returns the WorkerGroup component for starting all registered workers.
func (s *Server) Workers() *worker.WorkerGroup {
	return s.workerGroup
}

// Config returns the loaded application configuration.
func (s *Server) Config() *config.AppConfig {
	return s.cfg
}

// Shutdown cleanly closes all infrastructure connections.
// Workers are stopped by cancelling the context passed to Workers().StartGroup —
// call that cancel before Shutdown to ensure a clean drain sequence.
func (s *Server) Shutdown() {
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
	}

	if s.tracerStop != nil {
		if err := s.tracerStop(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", slog.String("error", err.Error()))
		}
	}

	if s.database != nil {
		if err := s.database.Close(); err != nil {
			slog.Error("failed to close database", slog.String("error", err.Error()))
		}
	}

	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			slog.Error("failed to close redis client", slog.String("error", err.Error()))
		}
	}
}
