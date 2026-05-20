package server

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/tdatIT/go-template/config"
	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
)

type fakeORM struct {
	closeCalled bool
	closeErr    error
}

func (f *fakeORM) GormDB() *gorm.DB { return nil }
func (f *fakeORM) SqlDB() *sql.DB   { return nil }
func (f *fakeORM) Close() error {
	f.closeCalled = true
	return f.closeErr
}

type fakeRedisClient struct {
	closeCalled bool
	closeErr    error
}

func (f *fakeRedisClient) Client() redis.UniversalClient { return nil }
func (f *fakeRedisClient) Close() error {
	f.closeCalled = true
	return f.closeErr
}

var _ orm.ORM = (*fakeORM)(nil)
var _ rdclient.RedisClient = (*fakeRedisClient)(nil)

func TestServerAccessorsAndShutdown(t *testing.T) {
	cfg := &config.AppConfig{Server: config.Server{Name: "svc"}}
	api := echo.New()
	db := &fakeORM{closeErr: errors.New("close db")}
	redis := &fakeRedisClient{closeErr: errors.New("close redis")}

	serv := Server{
		cfg:         cfg,
		echo:        api,
		database:    db,
		redisClient: redis,
	}

	require.Same(t, api, serv.API())
	require.Same(t, cfg, serv.Config())

	serv.Shutdown()
	require.True(t, db.closeCalled)
	require.True(t, redis.closeCalled)
}
