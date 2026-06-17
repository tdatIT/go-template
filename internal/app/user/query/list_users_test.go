package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tdatIT/go-template/internal/app/user/query"
	"github.com/tdatIT/go-template/internal/domain/dtos/userdtos"
	"github.com/tdatIT/go-template/internal/domain/models"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
	"github.com/tdatIT/go-template/pkgs/ultis/svcerr"
)

func TestListUsersQuery_Handle_Pagination(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		giveLimit  int
		giveOffset int
		wantLimit  int
		wantOffset int
	}{
		{name: "defaults when limit is zero", giveLimit: 0, giveOffset: 0, wantLimit: 50, wantOffset: 0},
		{name: "defaults when limit is negative", giveLimit: -5, giveOffset: 0, wantLimit: 50, wantOffset: 0},
		{name: "clamps limit above max", giveLimit: 500, giveOffset: 10, wantLimit: 200, wantOffset: 10},
		{name: "normalizes negative offset", giveLimit: 20, giveOffset: -3, wantLimit: 20, wantOffset: 0},
		{name: "passes valid values through", giveLimit: 25, giveOffset: 5, wantLimit: 25, wantOffset: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockRepository(t)
			repo.EXPECT().
				FindAndCount(mock.Anything, tt.wantLimit, tt.wantOffset).
				Return([]*models.User{}, int64(0), nil).
				Once()

			res, err := query.NewListUsersQuery(repo).
				Handle(ctx, &userdtos.ListUsersReq{Limit: tt.giveLimit, Offset: tt.giveOffset})

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

func TestListUsersQuery_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("success maps items and total", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		users := []*models.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		}
		repo.EXPECT().FindAndCount(mock.Anything, 50, 0).Return(users, int64(2), nil).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &userdtos.ListUsersReq{})

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, int64(2), res.Total)
		require.Len(t, res.Items, 2)
		assert.Equal(t, uint(1), res.Items[0].ID)
		assert.Equal(t, "Alice", res.Items[0].Name)
		assert.Equal(t, "Bob", res.Items[1].Name)
	})

	t.Run("empty result returns empty slice not nil", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindAndCount(mock.Anything, 50, 0).Return([]*models.User{}, int64(0), nil).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &userdtos.ListUsersReq{})

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, int64(0), res.Total)
		assert.NotNil(t, res.Items)
		assert.Empty(t, res.Items)
	})

	t.Run("repo error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindAndCount(mock.Anything, 50, 0).Return(nil, int64(0), errors.New("db down")).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &userdtos.ListUsersReq{})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}
