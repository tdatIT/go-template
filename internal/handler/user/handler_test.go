package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	userapp "github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	userrepo "github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
	"github.com/tdatIT/go-template/pkgs/ultis/validate"
)

type fakeRepo struct {
	findByIDFn     func(ctx context.Context, id uint) (*models.User, error)
	findAndCountFn func(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
	createFn       func(ctx context.Context, user *models.User) error
	updateFn       func(ctx context.Context, user *models.User) error
	deleteFn       func(ctx context.Context, id uint) error
}

func (f *fakeRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeRepo) FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	if f.findAndCountFn != nil {
		return f.findAndCountFn(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (f *fakeRepo) Create(ctx context.Context, user *models.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	return nil
}

func (f *fakeRepo) Update(ctx context.Context, user *models.User) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, user)
	}
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, id uint) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

var _ userrepo.Repository = (*fakeRepo)(nil)

type baseResPayload struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func newJSONContext(method, path, body string) (*echo.Echo, *echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = validate.GetValidator()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	return e, ctx, rec
}

func decodeBaseRes(t *testing.T, rec *httptest.ResponseRecorder) baseResPayload {
	var payload baseResPayload
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	return payload
}

func TestHandlerCreateUser(t *testing.T) {
	repo := &fakeRepo{
		createFn: func(ctx context.Context, user *models.User) error {
			user.ID = 1
			return nil
		},
	}
	app := userapp.NewUserApplication(repo)
	handler := NewUserHandler(app)

	_, ctx, rec := newJSONContext(http.MethodPost, "/api/v1/users", `{"name":"Alice","email":"alice@example.com"}`)
	require.NoError(t, handler.CreateUser(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	payload := decodeBaseRes(t, rec)
	require.Equal(t, "success", payload.Message)

	var dto userdtos.UserDTO
	require.NoError(t, json.Unmarshal(payload.Data, &dto))
	require.Equal(t, uint(1), dto.ID)
	require.Equal(t, "Alice", dto.Name)
}

func TestHandlerUpdateUser(t *testing.T) {
	repo := &fakeRepo{
		findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
			return &models.User{ID: id, Name: "Old", Email: "old@example.com"}, nil
		},
		updateFn: func(ctx context.Context, user *models.User) error {
			return nil
		},
	}
	app := userapp.NewUserApplication(repo)
	handler := NewUserHandler(app)

	_, ctx, rec := newJSONContext(http.MethodPut, "/api/v1/users/1", `{"name":"Bob","email":"bob@example.com"}`)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	require.NoError(t, handler.UpdateUser(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	payload := decodeBaseRes(t, rec)
	var dto userdtos.UserDTO
	require.NoError(t, json.Unmarshal(payload.Data, &dto))
	require.Equal(t, "Bob", dto.Name)
	require.Equal(t, "bob@example.com", dto.Email)
}

func TestHandlerDeleteUser(t *testing.T) {
	repo := &fakeRepo{
		deleteFn: func(ctx context.Context, id uint) error {
			return nil
		},
	}
	app := userapp.NewUserApplication(repo)
	handler := NewUserHandler(app)

	_, ctx, rec := newJSONContext(http.MethodDelete, "/api/v1/users/1", "")
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	require.NoError(t, handler.DeleteUser(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	payload := decodeBaseRes(t, rec)
	require.Equal(t, "success", payload.Message)
	require.Equal(t, "null", strings.TrimSpace(string(payload.Data)))
}

func TestHandlerGetUser(t *testing.T) {
	repo := &fakeRepo{
		findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
			return &models.User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
		},
	}
	app := userapp.NewUserApplication(repo)
	handler := NewUserHandler(app)

	_, ctx, rec := newJSONContext(http.MethodGet, "/api/v1/users/1", "")
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	require.NoError(t, handler.GetUser(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	payload := decodeBaseRes(t, rec)
	var dto userdtos.UserDTO
	require.NoError(t, json.Unmarshal(payload.Data, &dto))
	require.Equal(t, uint(1), dto.ID)
}

func TestHandlerListUsers(t *testing.T) {
	repo := &fakeRepo{
		findAndCountFn: func(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
			return []*models.User{
				{ID: 1, Name: "A", Email: "a@example.com"},
				{ID: 2, Name: "B", Email: "b@example.com"},
			}, 2, nil
		},
	}
	app := userapp.NewUserApplication(repo)
	handler := NewUserHandler(app)

	_, ctx, rec := newJSONContext(http.MethodGet, "/api/v1/users?limit=10&offset=0", "")
	require.NoError(t, handler.ListUsers(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	payload := decodeBaseRes(t, rec)
	var listRes userdtos.ListUsersRes
	require.NoError(t, json.Unmarshal(payload.Data, &listRes))
	require.Len(t, listRes.Items, 2)
	require.Equal(t, int64(2), listRes.Total)
}

func TestParseIDParam(t *testing.T) {
	_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users/1", "")
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})
	id, err := parseIDParam(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), id)

	_, ctx, _ = newJSONContext(http.MethodGet, "/api/v1/users/bad", "")
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	_, err = parseIDParam(ctx)
	require.Same(t, svcerr.ErrInvalidIdParam, err)
}

func TestParsePagination(t *testing.T) {
	_, ctx, _ := newJSONContext(http.MethodGet, "/api/v1/users?limit=5&offset=3", "")
	limit, offset, err := parsePagination(ctx)
	require.NoError(t, err)
	require.Equal(t, 5, limit)
	require.Equal(t, 3, offset)

	_, ctx, _ = newJSONContext(http.MethodGet, "/api/v1/users?limit=bad", "")
	_, _, err = parsePagination(ctx)
	require.Same(t, svcerr.ErrInvalidParameters, err)
}
