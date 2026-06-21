package orm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowThreshold = 200 * time.Millisecond

// slogGormLogger adapts an *slog.Logger to gorm's logger.Interface so GORM
// routes its output through the application's structured logging pipeline.
type slogGormLogger struct {
	logger        *slog.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(l *slog.Logger, level gormlogger.LogLevel, slowThreshold time.Duration) gormlogger.Interface {
	if slowThreshold <= 0 {
		slowThreshold = defaultSlowThreshold
	}
	return &slogGormLogger{
		logger:        l,
		level:         level,
		slowThreshold: slowThreshold,
	}
}

func (l *slogGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Info {
		return
	}
	l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Warn {
		return
	}
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Error {
		return
	}
	l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	attrs := []any{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gormlogger.ErrRecordNotFound):
		l.logger.ErrorContext(ctx, "gorm query failed", append(attrs, slog.String("error", err.Error()))...)
	case elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		l.logger.WarnContext(ctx, "gorm slow query", attrs...)
	case l.level >= gormlogger.Info:
		l.logger.InfoContext(ctx, "gorm query", attrs...)
	}
}
