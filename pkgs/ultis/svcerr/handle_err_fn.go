package svcerr

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/tdatIT/go-template/internal/domain/enums"
	"github.com/tdatIT/go-template/pkgs/ultis/response"
)

func ErrorHandlerEchoFn(ctx *echo.Context, err error) {
	msg := response.BaseRes{
		Status:  http.StatusInternalServerError,
		Code:    enums.Internal,
		Message: "internal server error",
	}

	if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		msg.Status = echoErr.Code
		msg.Code = echoErr.Code
		msg.Message = fmt.Sprintf("%v", echoErr.Message)
	}

	if customErr, ok := errors.AsType[*Error](err); ok {
		msg.Status = customErr.Status
		msg.Code = customErr.Code
		msg.Message = customErr.Message
	}

	if validateErr, ok := errors.AsType[validator.ValidationErrors](err); ok {
		msg.Status = http.StatusBadRequest
		msg.Code = enums.InvalidArgument
		if os.Getenv("SERVER_DEBUG") == "false" {
			msg.Message = "Invalid parameter"
		} else {
			msg.Message = validateErr.Error()

		}
	}

	if resp, uErr := echo.UnwrapResponse(ctx.Response()); uErr == nil {
		if !resp.Committed {
			ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			err = ctx.JSON(msg.Status, msg)
			if err != nil {
				ctx.Logger().Error("failed to send error response", slog.String("error", err.Error()))
			}
		}
	}
}
