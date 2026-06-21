package server

import (
	"context"
	"log/slog"
	"time"

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
	"github.com/tdatIT/go-template/pkgs/probe"
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

func NewServer() (*Server, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	slog.SetDefault(logger.NewSlogLogger(
		&logger.SlogConfig{
			Level:       cfg.Logger.Level,
			Format:      cfg.Logger.Format,
			ServiceName: cfg.Server.Name,
		},
		true,
	))

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
			return nil, traceErr
		}
		tracerStop = stop
	}

	database, err := orm.NewDBConnection(cfg)
	if err != nil {
		return nil, err
	}

	redisClient, redisErr := rdclient.NewRedisClient(cfg)
	if redisErr != nil {
		slog.Warn("redis unavailable at startup; readiness probe will report degraded",
			slog.String("error", redisErr.Error()))
	}

	readyProbe := probe.New(3*time.Second).
		Register("postgres", probe.DBChecker(database))
	if redisErr == nil {
		readyProbe.Register("redis", probe.RedisChecker(redisClient))
	} else {
		readyProbe.Register("redis", probe.CheckerFunc(func(_ context.Context) error {
			return redisErr
		}))
	}

	echoApp := newHttpServer(cfg)

	userRepository := userRepos.NewUserRepository(database)

	productAdapter := productsvc.NewAdapter(&cfg.Adapters.ProductService)

	userApplication := userApp.NewUserApplication(userRepository, productAdapter)

	usrHandle := userHandler.NewUserHandler(userApplication)

	// Register new router
	router.RegisterRoutes(
		echoApp,
		usrHandle,
		readyProbe,
	)

	// WorkerGroup is always created; workers are only registered when MQTT is available.
	workerGroup := worker.NewWorkerGroup()

	mqttCli, err := mqttpkg.NewMQTTClient(cfg)
	if err != nil {
		return nil, err
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
	}, nil
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
