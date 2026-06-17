package user_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/tdatIT/go-template/internal/domain/models"
	"github.com/tdatIT/go-template/internal/infras/repository/user"
)

// sqliteORM is a lightweight orm.ORM implementation backed by an in-process
// SQLite database, used to exercise the GORM repository without a real Postgres.
type sqliteORM struct{ db *gorm.DB }

func (s sqliteORM) GormDB() *gorm.DB { return s.db }
func (s sqliteORM) SqlDB() *sql.DB   { d, _ := s.db.DB(); return d }
func (s sqliteORM) Close() error     { d, _ := s.db.DB(); return d.Close() }

// newTestRepo spins up a fresh migrated SQLite DB and returns the repository
// under test plus the raw *gorm.DB for seeding/verification.
func newTestRepo(t *testing.T) (user.Repository, *gorm.DB) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gdb.AutoMigrate(&models.User{}))
	return user.NewUserRepository(sqliteORM{db: gdb}), gdb
}

// newBrokenRepo returns a repository whose backing DB has no `user` table
// (AutoMigrate skipped), so every query fails with a real SQL error. Used to
// cover the error-return branches of each method.
func newBrokenRepo(t *testing.T) user.Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "broken.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	return user.NewUserRepository(sqliteORM{db: gdb})
}

func seedUser(t *testing.T, gdb *gorm.DB, name, email string) *models.User {
	t.Helper()
	u := &models.User{Name: name, Email: email}
	require.NoError(t, gdb.Create(u).Error)
	return u
}

func TestUserRepository_Create(t *testing.T) {
	repo, gdb := newTestRepo(t)
	ctx := context.Background()

	u := &models.User{Name: "Alice", Email: "alice@example.com"}
	require.NoError(t, repo.Create(ctx, u))

	assert.NotZero(t, u.ID, "primary key should be populated after insert")

	var got models.User
	require.NoError(t, gdb.First(&got, u.ID).Error)
	assert.Equal(t, "Alice", got.Name)
	assert.Equal(t, "alice@example.com", got.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &models.User{Name: "Alice", Email: "dup@example.com"}))
	err := repo.Create(ctx, &models.User{Name: "Bob", Email: "dup@example.com"})

	assert.Error(t, err, "unique index on email must be enforced")
}

func TestUserRepository_FindByID(t *testing.T) {
	repo, gdb := newTestRepo(t)
	ctx := context.Background()
	seeded := seedUser(t, gdb, "Carol", "carol@example.com")

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, seeded.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, seeded.ID, got.ID)
		assert.Equal(t, "Carol", got.Name)
	})

	t.Run("not found returns gorm.ErrRecordNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 999)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestUserRepository_FindAndCount(t *testing.T) {
	repo, gdb := newTestRepo(t)
	ctx := context.Background()
	seedUser(t, gdb, "U1", "u1@example.com")
	seedUser(t, gdb, "U2", "u2@example.com")
	seedUser(t, gdb, "U3", "u3@example.com")

	t.Run("returns total alongside the page", func(t *testing.T) {
		items, total, err := repo.FindAndCount(ctx, 2, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, items, 2)
	})

	t.Run("respects offset", func(t *testing.T) {
		items, total, err := repo.FindAndCount(ctx, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, items, 1)
	})

	t.Run("empty table returns zero total and no rows", func(t *testing.T) {
		emptyRepo, _ := newTestRepo(t)
		items, total, err := emptyRepo.FindAndCount(ctx, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, items)
	})
}

func TestUserRepository_Update(t *testing.T) {
	repo, gdb := newTestRepo(t)
	ctx := context.Background()
	seeded := seedUser(t, gdb, "Old", "old@example.com")

	seeded.Name = "New"
	seeded.Email = "new@example.com"
	require.NoError(t, repo.Update(ctx, seeded))

	var got models.User
	require.NoError(t, gdb.First(&got, seeded.ID).Error)
	assert.Equal(t, "New", got.Name)
	assert.Equal(t, "new@example.com", got.Email)
}

func TestUserRepository_Delete(t *testing.T) {
	repo, gdb := newTestRepo(t)
	ctx := context.Background()

	t.Run("removes an existing row", func(t *testing.T) {
		seeded := seedUser(t, gdb, "ToDelete", "del@example.com")

		require.NoError(t, repo.Delete(ctx, seeded.ID))

		var count int64
		require.NoError(t, gdb.Model(&models.User{}).Where("id = ?", seeded.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("deleting a missing row returns gorm.ErrRecordNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, 999)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

// TestUserRepository_QueryErrors covers the error-return branches of every
// method by issuing queries against a DB with no `user` table.
func TestUserRepository_QueryErrors(t *testing.T) {
	repo := newBrokenRepo(t)
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		assert.Error(t, repo.Create(ctx, &models.User{Name: "X", Email: "x@example.com"}))
	})

	t.Run("find by id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, 1)
		assert.Nil(t, got)
		assert.Error(t, err)
	})

	t.Run("find and count", func(t *testing.T) {
		items, total, err := repo.FindAndCount(ctx, 10, 0)
		assert.Nil(t, items)
		assert.Zero(t, total)
		assert.Error(t, err)
	})

	t.Run("update", func(t *testing.T) {
		assert.Error(t, repo.Update(ctx, &models.User{ID: 1, Name: "X"}))
	})

	t.Run("delete", func(t *testing.T) {
		assert.Error(t, repo.Delete(ctx, 1))
	})
}
