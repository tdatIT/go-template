package server

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

const tracerName = "github.com/tdatIT/go-template/internal"

// tracingMiddleware creates a server-side span per request for echo/v5.
// It is the echo/v5 equivalent of the otelecho contrib middleware (which only
// supports echo/v4) and relies on the global TracerProvider set by pkgs/tracing.
func tracingMiddleware() echo.MiddlewareFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			route := c.RouteInfo().Path
			spanName := route
			if spanName == "" {
				spanName = req.Method // unmatched route (404/405): avoid high-cardinality raw path
			}

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(req.Method),
					semconv.URLPath(req.URL.Path),
					semconv.URLScheme(c.Scheme()),
					semconv.HTTPRoute(route),
					semconv.ServerAddress(req.Host),
					semconv.ClientAddress(c.RealIP()),
					semconv.UserAgentOriginal(req.UserAgent()),
				),
			)
			defer span.End()

			c.SetRequest(req.WithContext(ctx))

			err := next(c)

			status := responseStatus(c, err)
			span.SetAttributes(semconv.HTTPResponseStatusCode(status))
			if err != nil {
				span.RecordError(err)
			}
			// Per OTel HTTP conventions, server spans are errors only on 5xx.
			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(status))
			}

			return err
		}
	}
}

// responseStatus resolves the HTTP status code for the span. When the handler
// returns an error, Echo's HTTPErrorHandler writes the response above this
// middleware, so the status is derived from the error instead of the writer.
func responseStatus(c *echo.Context, err error) int {
	if err == nil {
		if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil && resp.Status != 0 {
			return resp.Status
		}
		return http.StatusOK
	}

	var svcErr *svcerr.Error
	if errors.As(err, &svcErr) {
		return svcErr.Status
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		return echoErr.Code
	}

	var validationErr validator.ValidationErrors
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
