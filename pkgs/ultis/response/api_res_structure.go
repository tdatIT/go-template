package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/tdatIT/go-template/internal/domain/enums"
)

type BaseRes struct {
	Status  int         `json:"-"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (g *BaseRes) JSON(c *echo.Context) error {
	return c.JSON(http.StatusOK, g)
}

var (
	SuccessRes = BaseRes{
		Status:  200,
		Code:    enums.Ok,
		Message: "success",
		Data:    nil,
	}

	ErrorRes = BaseRes{
		Status:  500,
		Code:    enums.Internal,
		Message: "internal server error",
		Data:    nil,
	}
)
