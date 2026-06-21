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
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
	"github.com/tdatIT/go-template/pkgs/utilities/svcerr"
)

func TestDeleteUserCommand_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("record not found maps to record-not-found", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().Delete(mock.Anything, uint(9)).Return(gorm.ErrRecordNotFound).Once()

		err := command.NewDeleteUserCommand(repo).Handle(ctx, &userdtos.DeleteUserReq{ID: 9})

		assert.ErrorIs(t, err, svcerr.ErrRecordNotFound)
	})

	t.Run("unexpected error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().Delete(mock.Anything, uint(5)).Return(errors.New("db down")).Once()

		err := command.NewDeleteUserCommand(repo).Handle(ctx, &userdtos.DeleteUserReq{ID: 5})

		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})

	t.Run("success returns nil", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().Delete(mock.Anything, uint(7)).Return(nil).Once()

		err := command.NewDeleteUserCommand(repo).Handle(ctx, &userdtos.DeleteUserReq{ID: 7})

		require.NoError(t, err)
	})
}
