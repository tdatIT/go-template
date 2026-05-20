package command

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

func TestCreateUserCommand(t *testing.T) {
	t.Run("invalid parameters", func(t *testing.T) {
		cmd := NewCreateUserCommand(&fakeRepo{})
		_, err := cmd.Handle(context.Background(), &userdtos.CreateUserReq{
			Name:  "  ",
			Email: "user@example.com",
		})
		require.Same(t, svcerr.ErrInvalidParameters, err)
	})

	t.Run("duplicate key", func(t *testing.T) {
		cmd := NewCreateUserCommand(&fakeRepo{
			createFn: func(ctx context.Context, user *models.User) error {
				return gorm.ErrDuplicatedKey
			},
		})
		_, err := cmd.Handle(context.Background(), &userdtos.CreateUserReq{
			Name:  "Alice",
			Email: "alice@example.com",
		})
		require.Same(t, svcerr.ErrAlreadyExists, err)
	})

	t.Run("success", func(t *testing.T) {
		var created *models.User
		cmd := NewCreateUserCommand(&fakeRepo{
			createFn: func(ctx context.Context, user *models.User) error {
				user.ID = 99
				created = user
				return nil
			},
		})
		dto, err := cmd.Handle(context.Background(), &userdtos.CreateUserReq{
			Name:  " Alice ",
			Email: " alice@example.com ",
		})
		require.NoError(t, err)
		require.NotNil(t, dto)
		require.Equal(t, uint(99), dto.ID)
		require.Equal(t, "Alice", dto.Name)
		require.Equal(t, "alice@example.com", dto.Email)
		require.Equal(t, "Alice", created.Name)
		require.Equal(t, "alice@example.com", created.Email)
	})
}

func TestUpdateUserCommand(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{})
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 0})
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("missing update fields", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{})
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1})
		require.Same(t, svcerr.ErrInvalidParameters, err)
	})

	t.Run("record not found", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		})
		name := "Bob"
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1, Name: &name})
		require.Same(t, svcerr.ErrRecordNotFound, err)
	})

	t.Run("find error", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return nil, errors.New("db error")
			},
		})
		name := "Bob"
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1, Name: &name})
		require.Same(t, svcerr.ErrInternalServer, err)
	})

	t.Run("invalid name", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return &models.User{ID: id, Name: "Old", Email: "old@example.com"}, nil
			},
		})
		name := " "
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1, Name: &name})
		require.Same(t, svcerr.ErrInvalidParameters, err)
	})

	t.Run("duplicate key", func(t *testing.T) {
		cmd := NewUpdateUserCommand(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return &models.User{ID: id, Name: "Old", Email: "old@example.com"}, nil
			},
			updateFn: func(ctx context.Context, user *models.User) error {
				return gorm.ErrDuplicatedKey
			},
		})
		name := "Bob"
		_, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1, Name: &name})
		require.Same(t, svcerr.ErrAlreadyExists, err)
	})

	t.Run("success", func(t *testing.T) {
		var updated *models.User
		cmd := NewUpdateUserCommand(&fakeRepo{
			findByIDFn: func(ctx context.Context, id uint) (*models.User, error) {
				return &models.User{ID: id, Name: "Old", Email: "old@example.com"}, nil
			},
			updateFn: func(ctx context.Context, user *models.User) error {
				updated = user
				return nil
			},
		})
		name := " Bob "
		email := " bob@example.com "
		dto, err := cmd.Handle(context.Background(), &userdtos.UpdateUserReq{ID: 1, Name: &name, Email: &email})
		require.NoError(t, err)
		require.Equal(t, "Bob", dto.Name)
		require.Equal(t, "bob@example.com", dto.Email)
		require.Equal(t, "Bob", updated.Name)
		require.Equal(t, "bob@example.com", updated.Email)
	})
}

func TestDeleteUserCommand(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		cmd := NewDeleteUserCommand(&fakeRepo{})
		err := cmd.Handle(context.Background(), &userdtos.DeleteUserReq{ID: 0})
		require.Same(t, svcerr.ErrInvalidIdParam, err)
	})

	t.Run("record not found", func(t *testing.T) {
		cmd := NewDeleteUserCommand(&fakeRepo{
			deleteFn: func(ctx context.Context, id uint) error {
				return gorm.ErrRecordNotFound
			},
		})
		err := cmd.Handle(context.Background(), &userdtos.DeleteUserReq{ID: 1})
		require.Same(t, svcerr.ErrRecordNotFound, err)
	})

	t.Run("delete error", func(t *testing.T) {
		cmd := NewDeleteUserCommand(&fakeRepo{
			deleteFn: func(ctx context.Context, id uint) error {
				return errors.New("delete failed")
			},
		})
		err := cmd.Handle(context.Background(), &userdtos.DeleteUserReq{ID: 1})
		require.Same(t, svcerr.ErrInternalServer, err)
	})

	t.Run("success", func(t *testing.T) {
		cmd := NewDeleteUserCommand(&fakeRepo{
			deleteFn: func(ctx context.Context, id uint) error {
				return nil
			},
		})
		require.NoError(t, cmd.Handle(context.Background(), &userdtos.DeleteUserReq{ID: 1}))
	})
}
