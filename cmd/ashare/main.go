package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := applog.Init(applog.Options{Level: cfg.Log.Level, Format: cfg.Log.Format}); err != nil {
		return err
	}

	if args[0] == "web" {
		return cmdWeb(cfg)
	}

	timeout := 30 * time.Second
	if args[0] == "ingest" {
		timeout = 6 * time.Hour
	}
	if args[0] == "backtest" {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch args[0] {
	case "ping":
		return cmdPing(ctx, cfg)
	case "migrate":
		return cmdMigrate(ctx, cfg)
	case "status":
		return cmdStatus(ctx, cfg)
	case "ingest":
		return cmdIngest(ctx, cfg, args[1:])
	case "show":
		return cmdShow(ctx, cfg, args[1:])
	case "stats":
		return cmdStats(ctx, cfg, args[1:])
	case "backtest":
		return cmdBacktest(ctx, cfg, args[1:])
	default:
		return fmt.Errorf("未知命令 %q，使用 ashare help 查看说明", args[0])
	}
}

func cmdPing(ctx context.Context, cfg config.Config) error {
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)
	slog.Info("postgresql ok", "component", "cli")
	return nil
}

func cmdMigrate(ctx context.Context, cfg config.Config) error {
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)
	if err := db.Migrate(ctx, gdb, migrations.FS); err != nil {
		return err
	}
	slog.Info("migrate done", "component", "cli")
	return nil
}

func cmdStatus(ctx context.Context, cfg config.Config) error {
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	type migrationRow struct {
		Version   string
		AppliedAt time.Time `gorm:"column:applied_at"`
	}
	var applied []migrationRow
	if err := gdb.WithContext(ctx).
		Raw(`SELECT version, applied_at FROM schema_migrations ORDER BY version`).
		Scan(&applied).Error; err != nil {
		return fmt.Errorf("尚未完成 migrate，或 schema_migrations 不存在: %w", err)
	}

	fmt.Println("已应用迁移:")
	if len(applied) == 0 {
		fmt.Println("  (无)")
	}
	for _, row := range applied {
		fmt.Printf("  - %s  (%s)\n", row.Version, row.AppliedAt.Format(time.RFC3339))
	}

	var nInstr, nCal, nBar int64
	if err := gdb.WithContext(ctx).Raw(`SELECT
		(SELECT COUNT(*) FROM instrument),
		(SELECT COUNT(*) FROM trade_calendar),
		(SELECT COUNT(*) FROM daily_bar)`).Row().Scan(&nInstr, &nCal, &nBar); err != nil {
		return err
	}
	fmt.Printf("instrument=%d  trade_calendar=%d  daily_bar=%d\n", nInstr, nCal, nBar)
	return nil
}

func printUsage() {
	fmt.Print(`ashare — 个人 A 股量化系统

用法:
  ashare ping      检查数据库连接
  ashare migrate   执行 SQL 迁移
  ashare status    查看迁移与表行数
  ashare ingest    从 Tushare 灌数（calendar / instruments / daily / sync）
  ashare show CODE [-start] [-end] [-qfq]
  ashare stats CODE [-start] [-end]
  ashare backtest CODE [-start] [-end] [-fast] [-slow] [-cash] [-out]
  ashare web       启动数据看板 API（配合 web/ 前端）
  ashare help      显示帮助

环境变量:
  DATABASE_URL     PostgreSQL 连接串
  TUSHARE_TOKEN    Tushare Pro token
  LOG_LEVEL        debug|info|warn|error（默认 info）
  LOG_FORMAT       text|json（默认 text）
`)
}
