package orm

import (
	"database/sql"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	DBType          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

	db, err := gorm.Open(dial, gConfig)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		db = db.Debug()
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

	// migration tables
	err = db.AutoMigrate()
	if err != nil {
		slog.Error("auto migrate failed", slog.String("error", err.Error()))
		return nil, err
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
