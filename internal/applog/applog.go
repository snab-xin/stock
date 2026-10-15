package applog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Options struct {
	Level  string // debug | info | warn | error
	Format string // text | json
	Out    io.Writer
}

func Init(opts Options) error {
	level, err := ParseLevel(opts.Level)
	if err != nil {
		return err
	}
	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = "text"
	}
	out := opts.Out
	if out == nil {
		out = os.Stderr
	}

	handlerOpts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(out, handlerOpts)
	case "json":
		h = slog.NewJSONHandler(out, handlerOpts)
	default:
		return fmt.Errorf("log.format 无效 %q，可选 text / json", opts.Format)
	}

	slog.SetDefault(slog.New(h))
	return nil
}

func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("log.level 无效 %q，可选 debug / info / warn / error", s)
	}
}

// With 给子系统加固定字段，例如 component=ingest。
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

func DurationMS(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// GormTraceLevel：debug 时跟踪每条 SQL，否则只记慢查询和错误。
func GormTraceLevel(configured string) slog.Level {
	lv, err := ParseLevel(configured)
	if err != nil || lv > slog.LevelDebug {
		return slog.LevelWarn
	}
	return slog.LevelDebug
}
