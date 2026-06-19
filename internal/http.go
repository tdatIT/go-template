package server

import (
	"context"
	"log/slog"
	"time"

	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	appmiddleware "github.com/tdatIT/go-template/internal/middleware"

	"github.com/tdatIT/go-template/config"
	slogConfig "github.com/tdatIT/go-template/pkgs/logger"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
	"github.com/tdatIT/go-template/pkgs/ultis/validate"
)

func newHttpServer(cfg *config.AppConfig) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(15 * time.Second))
	e.Use(middleware.RequestID())
	e.Validator = validate.GetValidator()

	e.HTTPErrorHandler = svcerr.ErrorHandlerEchoFn
	e.Logger = slog.New(
		slogConfig.NewJsonSlogHandler(
			&slogConfig.SlogConfig{ServiceName: cfg.Server.Name},
			false,
		))

	if cfg.Tracing.Enabled {
		e.Use(appmiddleware.TracingMiddleware())
	}

	if cfg.RateLimit.Enabled {
		store := middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      cfg.RateLimit.Rate,
				Burst:     cfg.RateLimit.Burst,
				ExpiresIn: cfg.RateLimit.ExpiresIn,
			},
		)
		e.Use(middleware.RateLimiter(store))
	}

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Skipper: func(c *echo.Context) bool {
			p := c.Request().URL.Path
			return p == "/metrics" || p == "/live" || p == "/ready"
		},
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:       true,
		LogURI:          true,
		HandleError:     true,
		LogLatency:      true,
		LogMethod:       true,
		LogURIPath:      true,
		LogRequestID:    true,
		LogResponseSize: true,
		LogUserAgent:    true,
		LogRemoteIP:     true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.Int("status", v.Status),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("path", v.URIPath),
				slog.Duration("latency", v.Latency),
				slog.Int64("response_size", v.ResponseSize),
				slog.String("request_id", v.RequestID),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("user_agent", v.UserAgent),
			}

			if v.Error == nil {
				slog.LogAttrs(context.Background(), slog.LevelInfo, "request success", attrs...)
			} else {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				slog.LogAttrs(context.Background(), slog.LevelError, "request error", attrs...)
			}

			return nil
		},
	}))

	return e
}
