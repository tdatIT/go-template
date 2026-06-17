package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/app/user/command"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

func ptr[T any](v T) *T { return &v }

func TestUpdateUserCommand_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("no fields to update is rejected before hitting the repo", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)

		got, err := command.NewUpdateUserCommand(repo).
			Handle(ctx, &userdtos.UpdateUserReq{ID: 1})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInvalidParameters)
		repo.AssertNotCalled(t, "FindByID")
	})

	t.Run("record not found on lookup maps to record-not-found", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindByID(mock.Anything, uint(9)).Return(nil, gorm.ErrRecordNotFound).Once()

		got, err := command.NewUpdateUserCommand(repo).
			Handle(ctx, &userdtos.UpdateUserReq{ID: 9, Name: ptr("Alice")})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrRecordNotFound)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("unexpected lookup error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindByID(mock.Anything, uint(9)).Return(nil, errors.New("db down")).Once()

		got, err := command.NewUpdateUserCommand(repo).
			Handle(ctx, &userdtos.UpdateUserReq{ID: 9, Name: ptr("Alice")})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
		repo.AssertNotCalled(t, "Update")
	})

	t.Run("duplicate key on update maps to already-exists", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindByID(mock.Anything, uint(1)).
			Return(&models.User{ID: 1, Name: "Old", Email: "old@example.com"}, nil).Once()
		repo.EXPECT().Update(mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey).Once()

		got, err := command.NewUpdateUserCommand(repo).
			Handle(ctx, &userdtos.UpdateUserReq{ID: 1, Email: ptr("dup@example.com")})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrAlreadyExists)
	})

	t.Run("unexpected update error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindByID(mock.Anything, uint(1)).
			Return(&models.User{ID: 1, Name: "Old", Email: "old@example.com"}, nil).Once()
		repo.EXPECT().Update(mock.Anything, mock.Anything).Return(errors.New("db down")).Once()

		got, err := command.NewUpdateUserCommand(repo).
			Handle(ctx, &userdtos.UpdateUserReq{ID: 1, Name: ptr("New")})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})

	t.Run("success applies trimmed name and email and returns dto", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindByID(mock.Anything, uint(1)).
			Return(&models.User{ID: 1, Name: "Old", Email: "old@example.com"}, nil).Once()
		repo.EXPECT().
			Update(mock.Anything, mock.MatchedBy(func(u *models.User) bool {
				return u.ID == 1 && u.Name == "New Name" && u.Email == "new@example.com"
			})).
			Return(nil).
			Once()

		got, err := command.NewUpdateUserCommand(repo).Handle(ctx, &userdtos.UpdateUserReq{
			ID:    1,
			Name:  ptr("  New Name  "),
			Email: ptr(" new@example.com "),
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(1), got.ID)
		assert.Equal(t, "New Name", got.Name)
		assert.Equal(t, "new@example.com", got.Email)
	})
}
