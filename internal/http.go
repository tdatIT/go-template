package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tdatIT/go-template/config"
	slogConfig "github.com/tdatIT/go-template/pkgs/logger"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
	"github.com/tdatIT/go-template/pkgs/ultis/validate"
)

func newHttpServer(cfg *config.AppConfig) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = svcerr.ErrorHandlerEchoFn
	e.Logger = slog.New(
		slogConfig.NewJsonSlogHandler(
			&slogConfig.SlogConfig{ServiceName: cfg.Server.Name},
			false,
		))
	e.Validator = validate.GetValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(15 * time.Second))
	e.Use(middleware.CSRF())
	e.Use(middleware.RequestID())

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
			// common attributes collected for every request log
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
				// append error message for error logs
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				slog.LogAttrs(context.Background(), slog.LevelError, "request error", attrs...)
			}

			return nil
		},
	}))

	return e
}
