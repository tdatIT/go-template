package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tdatIT/go-template/internal/app/user/query"
	"github.com/tdatIT/go-template/internal/domain/models"
	repomock "github.com/tdatIT/go-template/internal/infras/repository/user/mock"
	pagable "github.com/tdatIT/go-template/pkgs/utilities/paging"
	"github.com/tdatIT/go-template/pkgs/utilities/svcerr"
)

func TestListUsersQuery_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("success maps items and total", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		users := []*models.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		}
		// defaults: page=1, size=15 → FindAndCount(size=15, offset=0)
		repo.EXPECT().FindAndCount(mock.Anything, 15, 0).Return(users, int64(2), nil).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &pagable.ListQuery{})

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, 2, res.Total)
		items := res.Items.([]*models.User)
		require.Len(t, items, 2)
		assert.Equal(t, uint(1), items[0].ID)
		assert.Equal(t, "Alice", items[0].Name)
		assert.Equal(t, "Bob", items[1].Name)
	})

	t.Run("passes page and size to repo correctly", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		// page=2, size=10 → offset=(2-1)*10=10
		repo.EXPECT().FindAndCount(mock.Anything, 10, 10).Return([]*models.User{}, int64(25), nil).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &pagable.ListQuery{Page: 2, Size: 10})

		require.NoError(t, err)
		assert.Equal(t, 25, res.Total)
		assert.Equal(t, 2, res.Page)
		assert.Equal(t, 10, res.Size)
	})

	t.Run("empty result returns empty slice", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindAndCount(mock.Anything, 15, 0).Return([]*models.User{}, int64(0), nil).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &pagable.ListQuery{})

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, 0, res.Total)
		items := res.Items.([]*models.User)
		assert.Empty(t, items)
	})

	t.Run("repo error maps to internal server", func(t *testing.T) {
		repo := repomock.NewMockRepository(t)
		repo.EXPECT().FindAndCount(mock.Anything, 15, 0).Return(nil, int64(0), errors.New("db down")).Once()

		res, err := query.NewListUsersQuery(repo).Handle(ctx, &pagable.ListQuery{})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, svcerr.ErrInternalServer)
	})
}
