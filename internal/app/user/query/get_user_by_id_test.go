package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/app/user/query"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

func TestGetUserByIDQuery_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns dto", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			FindByID(mock.Anything, uint(7)).
			Return(&models.User{ID: 7, Name: "Carol", Email: "carol@example.com"}, nil).
			Once()

		q := query.NewGetUserByIDQuery(repo)
		got, err := q.Handle(ctx, &userdtos.GetUserByIDReq{ID: 7})

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(7), got.ID)
		assert.Equal(t, "Carol", got.Name)
	})

	t.Run("zero id is rejected before hitting the repo", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)

		q := query.NewGetUserByIDQuery(repo)
		got, err := q.Handle(ctx, &userdtos.GetUserByIDReq{ID: 0})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInvalidIdParam)
		repo.AssertNotCalled(t, "FindByID")
	})

	t.Run("record not found maps to record-not-found", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			FindByID(mock.Anything, uint(99)).
			Return(nil, gorm.ErrRecordNotFound).
			Once()

		q := query.NewGetUserByIDQuery(repo)
		got, err := q.Handle(ctx, &userdtos.GetUserByIDReq{ID: 99})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrRecordNotFound)
	})

	t.Run("unexpected error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().
			FindByID(mock.Anything, uint(5)).
			Return(nil, errors.New("connection refused")).
			Once()

		q := query.NewGetUserByIDQuery(repo)
		got, err := q.Handle(ctx, &userdtos.GetUserByIDReq{ID: 5})

		assert.Nil(t, got)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}
