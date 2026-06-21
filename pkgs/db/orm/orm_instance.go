package orm

import (
	"database/sql"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ORM defines a interface for access the db.
type ORM interface {
	GormDB() *gorm.DB
	SqlDB() *sql.DB
	Close() error
}

// Config GORM Config
type Config struct {
	Debug           bool
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	// Logger, when set, routes GORM logs through the given slog logger.
	Logger *slog.Logger
	// SlowThreshold flags queries slower than this as slow (0 = default 200ms).
	SlowThreshold time.Duration
}

type ormImpl struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func newGormInstance(c Config) (ORM, error) {
	dial := postgres.Open(c.DSN)
	gConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	if c.Logger != nil {
		level := gormlogger.Warn
		if c.Debug {
			level = gormlogger.Info
		}
		gConfig.Logger = newGormLogger(c.Logger, level, c.SlowThreshold)
	}

	db, err := gorm.Open(dial, gConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if c.MaxOpenConns != 0 {
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	}

	if c.ConnMaxLifetime != 0 {
		sqlDB.SetConnMaxLifetime(c.ConnMaxLifetime)
	}

	if c.MaxIdleConns != 0 {
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.ConnMaxIdleTime != 0 {
		sqlDB.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	}

	return &ormImpl{
		db:    db,
		sqlDB: sqlDB,
	}, nil
}

func (g *ormImpl) SqlDB() *sql.DB {
	return g.sqlDB
}

func (g *ormImpl) GormDB() *gorm.DB {
	return g.db
}

func (g *ormImpl) Close() error {
	return g.sqlDB.Close()
}
