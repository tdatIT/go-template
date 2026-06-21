package router

import (
	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"

	"github.com/tdatIT/go-template/internal/handler/health"
	"github.com/tdatIT/go-template/internal/handler/user"
)

func RegisterRoutes(e *echo.Echo,
	userHandler *user.Handler,
	healthHandler *health.Handler,
) {
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/liveness", healthHandler.Live)
	e.GET("/readiness", healthHandler.Ready)

	api := e.Group("/api")
	v1 := api.Group("/v1")

	userRoute := v1.Group("/users")
	userRoute.POST("", userHandler.CreateUser)
	userRoute.GET("", userHandler.ListUsers)
	userRoute.GET("/:id", userHandler.GetUser)
	userRoute.PUT("/:id", userHandler.UpdateUser)
	userRoute.DELETE("/:id", userHandler.DeleteUser)
}
