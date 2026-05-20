package router

import (
	"context"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	userapp "github.com/tdatIT/go-template/internal/app/user"
	"github.com/tdatIT/go-template/internal/domain/models"
	userhandler "github.com/tdatIT/go-template/internal/handler/user"
	userrepo "github.com/tdatIT/go-template/internal/infras/repository/user"
)

type fakeRepo struct{}

func (f *fakeRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	return &models.User{ID: id}, nil
}
func (f *fakeRepo) FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	return []*models.User{}, 0, nil
}
func (f *fakeRepo) Create(ctx context.Context, user *models.User) error { return nil }
func (f *fakeRepo) Update(ctx context.Context, user *models.User) error { return nil }
func (f *fakeRepo) Delete(ctx context.Context, id uint) error           { return nil }

var _ userrepo.Repository = (*fakeRepo)(nil)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	app := userapp.NewUserApplication(&fakeRepo{})
	handler := userhandler.NewUserHandler(app)

	RegisterRoutes(e, handler)
	routes := e.Router().Routes()

	_, err := routes.FindByMethodPath(http.MethodPost, "/api/v1/users")
	require.NoError(t, err)
	_, err = routes.FindByMethodPath(http.MethodGet, "/api/v1/users")
	require.NoError(t, err)
	_, err = routes.FindByMethodPath(http.MethodGet, "/api/v1/users/:id")
	require.NoError(t, err)
	_, err = routes.FindByMethodPath(http.MethodPut, "/api/v1/users/:id")
	require.NoError(t, err)
	_, err = routes.FindByMethodPath(http.MethodDelete, "/api/v1/users/:id")
	require.NoError(t, err)
}
