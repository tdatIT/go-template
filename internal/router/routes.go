package router

import (
	"net/http"

	echoprometheus "github.com/labstack/echo-prometheus"
	"github.com/labstack/echo/v5"

	"github.com/tdatIT/go-template/internal/handler/user"
	"github.com/tdatIT/go-template/pkgs/probe"
)

func RegisterRoutes(e *echo.Echo,
	userHandler *user.Handler,
	readyProbe *probe.Probe,
) {
	// metrics router
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/liveness", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/readiness", echo.WrapHandler(readyProbe.Handler()))

	// init default router
	api := e.Group("/api")
	v1 := api.Group("/v1")

	userRoute := v1.Group("/users")
	userRoute.POST("", userHandler.CreateUser)
	userRoute.GET("", userHandler.ListUsers)
	userRoute.GET("/:id", userHandler.GetUser)
	userRoute.PUT("/:id", userHandler.UpdateUser)
	userRoute.DELETE("/:id", userHandler.DeleteUser)
}
