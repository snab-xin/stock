package applog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGormLogger 把 GORM 慢查询/错误接到 slog。默认只记 Warn 及以上（慢 SQL、出错）。
func NewGormLogger(level slog.Level) gormlogger.Interface {
	return &gormSlog{
		level:         level,
		slowThreshold: 200 * time.Millisecond,
	}
}

type gormSlog struct {
	level         slog.Level
	slowThreshold time.Duration
}

func (l *gormSlog) LogMode(gormlogger.LogLevel) gormlogger.Interface { return l }

func (l *gormSlog) Info(ctx context.Context, msg string, data ...any) {
	if l.level > slog.LevelInfo {
		return
	}
	slog.InfoContext(ctx, fmt.Sprintf(msg, data...), "component", "gorm")
}

func (l *gormSlog) Warn(ctx context.Context, msg string, data ...any) {
	if l.level > slog.LevelWarn {
		return
	}
	slog.WarnContext(ctx, fmt.Sprintf(msg, data...), "component", "gorm")
}

func (l *gormSlog) Error(ctx context.Context, msg string, data ...any) {
	slog.ErrorContext(ctx, fmt.Sprintf(msg, data...), "component", "gorm")
}

func (l *gormSlog) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		slog.ErrorContext(ctx, "sql error",
			"component", "gorm",
			"elapsed_ms", DurationMS(elapsed),
			"rows", rows,
			"err", err,
			"sql", sql,
		)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level <= slog.LevelWarn:
		slog.WarnContext(ctx, "slow sql",
			"component", "gorm",
			"elapsed_ms", DurationMS(elapsed),
			"rows", rows,
			"sql", sql,
		)
	case l.level <= slog.LevelDebug:
		slog.DebugContext(ctx, "sql",
			"component", "gorm",
			"elapsed_ms", DurationMS(elapsed),
			"rows", rows,
			"sql", sql,
		)
	}
}
