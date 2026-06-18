package user

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	userapp "github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/domain/models"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

func newHandler(repo *fakeRepo) *Handler {
	return NewUserHandler(userapp.NewUserApplication(repo, &fakeProductAdapter{}))
}

func TestHandlerCreateUser_Errors(t *testing.T) {
	t.Run("malformed json body returns bad request", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodPost, "/api/v1/users", `{"name":`)
		err := h.CreateUser(ctx)
		require.Same(t, svcerr.ErrBadRequest, err)
	})

	t.Run("validation failure is returned", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodPost, "/api/v1/users", `{"name":"A","email":"not-an-email"}`)
		err := h.CreateUser(ctx)
		require.Error(t, err)
	})

	t.Run("application error is propagated", func(t *testing.T) {
		h := newHandler(&fakeRepo{
			createFn: func(context.Context, *models.User) error { return gorm.ErrDuplicatedKey },
		})
		_, ctx, _ := newJSONContext(http.MethodPost, "/api/v1/users", `{"name":"Alice","email":"alice@example.com"}`)
		err := h.CreateUser(ctx)
		assert.ErrorIs(t, err, svcerr.ErrAlreadyExists)
	})
}

func TestHandlerUpdateUser_Errors(t *testing.T) {
	t.Run("invalid id param", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodPut, "/api/v1/users/bad", `{"name":"Bob"}`)
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
		err := h.UpdateUser(ctx)
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("malformed json body returns bad request", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodPut, "/api/v1/users/1", `{"name":`)
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
		err := h.UpdateUser(ctx)
		require.Same(t, svcerr.ErrBadRequest, err)
	})

	t.Run("validation failure is returned", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodPut, "/api/v1/users/1", `{"email":"not-an-email"}`)
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
		err := h.UpdateUser(ctx)
		require.Error(t, err)
	})

	t.Run("application error is propagated", func(t *testing.T) {
		h := newHandler(&fakeRepo{
			findByIDFn: func(context.Context, uint) (*models.User, error) { return nil, gorm.ErrRecordNotFound },
		})
		_, ctx, _ := newJSONContext(http.MethodPut, "/api/v1/users/1", `{"name":"Bob","email":"bob@example.com"}`)
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
		err := h.UpdateUser(ctx)
		assert.ErrorIs(t, err, svcerr.ErrRecordNotFound)
	})
}

func TestHandlerDeleteUser_Errors(t *testing.T) {
	t.Run("invalid id param", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodDelete, "/api/v1/users/bad", "")
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
		err := h.DeleteUser(ctx)
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("application error is propagated", func(t *testing.T) {
		h := newHandler(&fakeRepo{
			deleteFn: func(context.Context, uint) error { return errors.New("db down") },
		})
		_, ctx, _ := newJSONContext(http.MethodDelete, "/api/v1/users/1", "")
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
		err := h.DeleteUser(ctx)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}

func TestHandlerGetUser_Errors(t *testing.T) {
	t.Run("invalid id param", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users/bad", "")
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
		err := h.GetUser(ctx)
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("application error is propagated", func(t *testing.T) {
		h := newHandler(&fakeRepo{
			findByIDFn: func(context.Context, uint) (*models.User, error) { return nil, gorm.ErrRecordNotFound },
		})
		_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users/1", "")
		ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
		err := h.GetUser(ctx)
		assert.ErrorIs(t, err, svcerr.ErrRecordNotFound)
	})
}

func TestHandlerListUsers_Errors(t *testing.T) {
	t.Run("invalid pagination param", func(t *testing.T) {
		h := newHandler(&fakeRepo{})
		_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users?page=bad", "")
		err := h.ListUsers(ctx)
		require.Same(t, svcerr.ErrInvalidParameters, err)
	})

	t.Run("application error is propagated", func(t *testing.T) {
		h := newHandler(&fakeRepo{
			findAndCountFn: func(context.Context, int, int) ([]*models.User, int64, error) {
				return nil, 0, errors.New("db down")
			},
		})
		_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users", "")
		err := h.ListUsers(ctx)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}
