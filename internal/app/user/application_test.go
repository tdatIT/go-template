package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tdatIT/go-template/internal/domain/models"
	userrepo "github.com/tdatIT/go-template/internal/infras/repository/user"
)

type noopRepo struct{}

func (n *noopRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	return &models.User{ID: id}, nil
}
func (n *noopRepo) FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	return []*models.User{}, 0, nil
}
func (n *noopRepo) Create(ctx context.Context, user *models.User) error { return nil }
func (n *noopRepo) Update(ctx context.Context, user *models.User) error { return nil }
func (n *noopRepo) Delete(ctx context.Context, id uint) error           { return nil }

var _ userrepo.Repository = (*noopRepo)(nil)

func TestNewUserApplication(t *testing.T) {
	app := NewUserApplication(&noopRepo{})
	require.NotNil(t, app)
	require.NotNil(t, app.Command.CreateUser)
	require.NotNil(t, app.Command.UpdateUser)
	require.NotNil(t, app.Command.DeleteUser)
	require.NotNil(t, app.Query.GetUser)
	require.NotNil(t, app.Query.ListUsers)
}
