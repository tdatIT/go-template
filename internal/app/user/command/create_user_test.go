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

func TestCreateUserCommand_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("success trims input and returns dto", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(u *models.User) bool {
				// Whitespace must be trimmed before persisting.
				return u.Name == "Alice" && u.Email == "alice@example.com"
			})).
			Run(func(_ context.Context, u *models.User) {
				u.ID = 1 // simulate DB assigning a primary key
			}).
			Return(nil).
			Once()

		cmd := command.NewCreateUserCommand(repo)
		got, err := cmd.Handle(ctx, &userdtos.CreateUserReq{
			Name:  "  Alice  ",
			Email: " alice@example.com ",
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(1), got.ID)
		assert.Equal(t, "Alice", got.Name)
		assert.Equal(t, "alice@example.com", got.Email)
	})

	t.Run("empty name is rejected before hitting the repo", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)

		cmd := command.NewCreateUserCommand(repo)
		got, err := cmd.Handle(ctx, &userdtos.CreateUserReq{
			Name:  "   ",
			Email: "alice@example.com",
		})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInvalidParameters)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("duplicate key maps to already-exists", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(gorm.ErrDuplicatedKey).
			Once()

		cmd := command.NewCreateUserCommand(repo)
		got, err := cmd.Handle(ctx, &userdtos.CreateUserReq{Name: "Bob", Email: "bob@example.com"})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrAlreadyExists)
	})

	t.Run("unexpected error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(errors.New("connection refused")).
			Once()

		cmd := command.NewCreateUserCommand(repo)
		got, err := cmd.Handle(ctx, &userdtos.CreateUserReq{Name: "Bob", Email: "bob@example.com"})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}
