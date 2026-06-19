package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
)

type Handler struct {
	db    orm.ORM
	redis rdclient.RedisClient
}

func NewHealthHandler(db orm.ORM, redis rdclient.RedisClient) *Handler {
	return &Handler{db: db, redis: redis}
}

func (h *Handler) Live(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	details := map[string]string{}
	healthy := true

	if err := h.db.SqlDB().PingContext(ctx); err != nil {
		details["database"] = err.Error()
		healthy = false
	} else {
		details["database"] = "ok"
	}

	if err := h.redis.Client().Ping(ctx).Err(); err != nil {
		details["redis"] = err.Error()
		healthy = false
	} else {
		details["redis"] = "ok"
	}

	if !healthy {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"status":  "degraded",
			"details": details,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "ok",
		"details": details,
	})
}
