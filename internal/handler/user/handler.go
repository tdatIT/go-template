package user

import (
	"log/slog"
	"strconv"

	"github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	pagable "github.com/tdatIT/go-template/pkgs/utilities/paging"
	"github.com/tdatIT/go-template/pkgs/utilities/response"
	"github.com/tdatIT/go-template/pkgs/utilities/svcerr"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	app *user.Application
}

func NewUserHandler(app *user.Application) *Handler {
	return &Handler{app: app}
}

func (h *Handler) CreateUser(c *echo.Context) error {
	var req userdtos.CreateUserReq
	if err := c.Bind(&req); err != nil {
		slog.Error("bind request failed", slog.String("error", err.Error()))
		return svcerr.ErrBadRequest
	}

	if err := c.Validate(&req); err != nil {
		slog.Error("validate request failed", slog.String("error", err.Error()))
		return err
	}

	data, err := h.app.Command.CreateUser.Handle(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	success := response.SuccessRes
	success.Data = data

	return success.JSON(c)
}

func (h *Handler) UpdateUser(c *echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return err
	}

	var req userdtos.UpdateUserReq
	if err := c.Bind(&req); err != nil {
		slog.Error("bind request failed", slog.String("error", err.Error()))
		return svcerr.ErrBadRequest
	}

	if err := c.Validate(&req); err != nil {
		slog.Error("validate request failed", slog.String("error", err.Error()))
		return err
	}

	req.ID = id
	data, err := h.app.Command.UpdateUser.Handle(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	success := response.SuccessRes
	success.Data = data

	return success.JSON(c)
}

func (h *Handler) DeleteUser(c *echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return err
	}

	req := userdtos.DeleteUserReq{ID: id}
	if err := h.app.Command.DeleteUser.Handle(c.Request().Context(), &req); err != nil {
		return err
	}

	success := response.SuccessRes
	success.Data = nil

	return success.JSON(c)
}

func (h *Handler) GetUser(c *echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return err
	}

	req := userdtos.GetUserByIDReq{ID: id}
	data, err := h.app.Query.GetUser.Handle(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	success := response.SuccessRes
	success.Data = data

	return success.JSON(c)
}

func (h *Handler) ListUsers(c *echo.Context) error {
	q, err := parseListQuery(c)
	if err != nil {
		return err
	}

	data, err := h.app.Query.ListUsers.Handle(c.Request().Context(), q)
	if err != nil {
		return err
	}

	success := response.SuccessRes
	success.Data = data

	return success.JSON(c)
}

func parseIDParam(c *echo.Context) (uint, error) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		return 0, svcerr.ErrInvalidIdParam
	}
	return uint(id), nil
}

func parseListQuery(c *echo.Context) (*pagable.ListQuery, error) {
	var q pagable.ListQuery
	if err := q.SetPage(c.QueryParam("page")); err != nil {
		return nil, svcerr.ErrInvalidParameters
	}
	if err := q.SetSize(c.QueryParam("size")); err != nil {
		return nil, svcerr.ErrInvalidParameters
	}
	return &q, nil
}
