package user

import (
	"context"

	"github.com/tdatIT/go-template/internal/domain/models"
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindAndCount(ctx context.Context, limit, offset int) ([]*models.User, int64, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
}
