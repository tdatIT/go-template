package query

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
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

var _ user.Repository = (*fakeRepo)(nil)

func TestGetUserByIDQuery(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		q := NewGetUserByIDQuery(&fakeRepo{})
		_, err := q.Handle(context.Background(), &userdtos.GetUserByIDReq{ID: 0})
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("record not found", func(t *testing.T) {
		q := NewGetUserByIDQuery(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		})
		_, err := q.Handle(context.Background(), &userdtos.GetUserByIDReq{ID: 1})
		require.Same(t, svcerr.ErrRecordNotFound, err)
	})

	t.Run("find error", func(t *testing.T) {
		q := NewGetUserByIDQuery(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return nil, errors.New("find failed")
			},
		})
		_, err := q.Handle(context.Background(), &userdtos.GetUserByIDReq{ID: 1})
		require.Same(t, svcerr.ErrInternalServer, err)
	})

	t.Run("success", func(t *testing.T) {
		q := NewGetUserByIDQuery(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return &models.User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
			},
		})
		dto, err := q.Handle(context.Background(), &userdtos.GetUserByIDReq{ID: 1})
		require.NoError(t, err)
		require.Equal(t, uint(1), dto.ID)
	})
}

func TestListUsersQuery(t *testing.T) {
	t.Run("default limit and offset", func(t *testing.T) {
		var gotLimit, gotOffset int
		q := NewListUsersQuery(&fakeRepo{
			findAndCountFn: func(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
				gotLimit = limit
				gotOffset = offset
				return []*models.User{}, 0, nil
			},
		})
		_, err := q.Handle(context.Background(), &userdtos.ListUsersReq{Limit: 0, Offset: -1})
		require.NoError(t, err)
		require.Equal(t, 50, gotLimit)
		require.Equal(t, 0, gotOffset)
	})

	t.Run("max limit clamp", func(t *testing.T) {
		var gotLimit int
		q := NewListUsersQuery(&fakeRepo{
			findAndCountFn: func(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
				gotLimit = limit
				return []*models.User{}, 0, nil
			},
		})
		_, err := q.Handle(context.Background(), &userdtos.ListUsersReq{Limit: 999, Offset: 0})
		require.NoError(t, err)
		require.Equal(t, 200, gotLimit)
	})

	t.Run("repo error", func(t *testing.T) {
		q := NewListUsersQuery(&fakeRepo{
			findAndCountFn: func(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
				return nil, 0, errors.New("list failed")
			},
		})
		_, err := q.Handle(context.Background(), &userdtos.ListUsersReq{Limit: 10, Offset: 0})
		require.Same(t, svcerr.ErrInternalServer, err)
	})
}
