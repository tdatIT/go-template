package router

import (
	"github.com/labstack/echo/v5"
	"github.com/tdatIT/go-template/internal/handler/user"
)

func RegisterRoutes(e *echo.Echo,
	userHandler *user.Handler,
) {
	api := e.Group("/api")
	v1 := api.Group("/v1")

	userRoute := v1.Group("/users")
	userRoute.POST("", userHandler.CreateUser)
	userRoute.GET("", userHandler.ListUsers)
	userRoute.GET("/:id", userHandler.GetUser)
	userRoute.PUT("/:id", userHandler.UpdateUser)
	userRoute.DELETE("/:id", userHandler.DeleteUser)
}
