package probe

import (
	"context"

	"github.com/tdatIT/go-template/pkgs/db/orm"
	"github.com/tdatIT/go-template/pkgs/db/rdclient"
)

// DBChecker returns a Checker that pings the relational database.
func DBChecker(db orm.ORM) CheckerFunc {
	return func(ctx context.Context) error {
		return db.SqlDB().PingContext(ctx)
	}
}

// RedisChecker returns a Checker that pings Redis.
func RedisChecker(rc rdclient.RedisClient) CheckerFunc {
	return func(ctx context.Context) error {
		return rc.Client().Ping(ctx).Err()
	}
}
