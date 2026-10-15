package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/stock/ashare/internal/app"
	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
)

func cmdIngest(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) == 0 {
		printIngestUsage()
		return nil
	}
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	svc, err := app.NewIngestService(cfg, gdb)
	if err != nil {
		return err
	}
	slog.Info("tushare rate limit",
		"component", "ingest",
		"rpm", cfg.TushareRPM,
		"apis", cfg.TushareAPIs,
		"gap", cfg.MinInterval().Round(time.Millisecond).String(),
	)

	switch args[0] {
	case "help", "-h", "--help":
		printIngestUsage()
		return nil
	case "instruments":
		return svc.SyncInstruments(ctx)
	case "calendar":
		start, end, _, _, _, err := parseIngestFlags(args[1:], "20100101", dates.YYYYMMDD(time.Now()))
		if err != nil {
			return err
		}
		return svc.SyncCalendar(ctx, start, end)
	case "daily":
		start, end, codes, _, _, err := parseIngestFlags(args[1:], "", "")
		if err != nil {
			return err
		}
		if start.IsZero() || end.IsZero() {
			return fmt.Errorf("daily 需要 -start 和 -end，例如 -start 20260101 -end 20261013")
		}
		resolved, err := resolveCodes(ctx, repo.NewInstrumentRepo(gdb), codes)
		if err != nil {
			return err
		}
		return svc.SyncDaily(ctx, start, end, resolved)
	case "sync":
		start, end, codes, _, _, err := parseIngestFlags(args[1:], "", "")
		if err != nil {
			return err
		}
		if start.IsZero() || end.IsZero() {
			return fmt.Errorf("sync 需要 -start 和 -end")
		}
		if err := svc.SyncCalendar(ctx, start, end); err != nil {
			return err
		}
		if err := svc.SyncInstruments(ctx); err != nil {
			return err
		}
		resolved, err := resolveCodes(ctx, repo.NewInstrumentRepo(gdb), codes)
		if err != nil {
			return err
		}
		return svc.SyncDaily(ctx, start, end, resolved)
	case "cb-basic":
		return svc.SyncCBBasic(ctx)
	case "cb-holders":
		start, end, codes, period, _, err := parseIngestFlags(args[1:], "", "")
		if err != nil {
			return err
		}
		if period.IsZero() && (start.IsZero() || end.IsZero()) {
			return fmt.Errorf("cb-holders 需要 -period 20250630，或 -start/-end 报告期区间")
		}
		for i := range codes {
			codes[i] = strings.ToUpper(strings.TrimSpace(codes[i]))
		}
		return svc.SyncCBTopHolders(ctx, period, start, end, codes)
	case "cb-daily":
		start, end, codes, _, tradeDate, err := parseIngestFlags(args[1:], "", "")
		if err != nil {
			return err
		}
		if tradeDate.IsZero() && (start.IsZero() || end.IsZero()) {
			return fmt.Errorf("cb-daily 需要 -date 20261015，或 -start/-end；可加 -code")
		}
		for i := range codes {
			codes[i] = strings.ToUpper(strings.TrimSpace(codes[i]))
		}
		return svc.SyncCBDaily(ctx, start, end, tradeDate, codes)
	default:
		return fmt.Errorf("未知 ingest 子命令 %q", args[0])
	}
}

func parseIngestFlags(args []string, defaultStart, defaultEnd string) (start, end time.Time, codes []string, period, tradeDate time.Time, err error) {
	fs := flag.NewFlagSet("ingest", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	startStr := fs.String("start", defaultStart, "开始日期 YYYYMMDD")
	endStr := fs.String("end", defaultEnd, "结束日期 YYYYMMDD")
	periodStr := fs.String("period", "", "报告期 YYYYMMDD（可转债十大持有人）")
	dateStr := fs.String("date", "", "交易日 YYYYMMDD（可转债行情截面）")
	codeStr := fs.String("code", "", "证券代码，逗号分隔，缺省为全市场")
	if err := fs.Parse(args); err != nil {
		return time.Time{}, time.Time{}, nil, time.Time{}, time.Time{}, err
	}
	if strings.TrimSpace(*startStr) != "" {
		start, err = dates.Parse(*startStr)
		if err != nil {
			return time.Time{}, time.Time{}, nil, time.Time{}, time.Time{}, err
		}
	}
	if strings.TrimSpace(*endStr) != "" {
		end, err = dates.Parse(*endStr)
		if err != nil {
			return time.Time{}, time.Time{}, nil, time.Time{}, time.Time{}, err
		}
	}
	if strings.TrimSpace(*periodStr) != "" {
		period, err = dates.Parse(*periodStr)
		if err != nil {
			return time.Time{}, time.Time{}, nil, time.Time{}, time.Time{}, err
		}
	}
	if strings.TrimSpace(*dateStr) != "" {
		tradeDate, err = dates.Parse(*dateStr)
		if err != nil {
			return time.Time{}, time.Time{}, nil, time.Time{}, time.Time{}, err
		}
	}
	if *codeStr != "" {
		for _, c := range strings.Split(*codeStr, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				codes = append(codes, c)
			}
		}
	}
	return start, end, codes, period, tradeDate, nil
}

func resolveCodes(ctx context.Context, instruments *repo.InstrumentRepo, codes []string) ([]string, error) {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		code, err := ingest.ResolveTSCode(ctx, instruments, c)
		if err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, nil
}

func printIngestUsage() {
	fmt.Print(`ashare ingest — 从 Tushare 灌数

用法:
  ashare ingest calendar [-start 20100101] [-end 20261013]
  ashare ingest instruments
  ashare ingest daily -start 20260101 -end 20261013 [-code 000001.SZ]
  ashare ingest sync  -start 20260101 -end 20261013 [-code 000001.SZ]
  ashare ingest cb-basic
  ashare ingest cb-daily -date 20261015
  ashare ingest cb-daily -start 20261001 -end 20261015 [-code 110059.SH]
  ashare ingest cb-holders -period 20250630 [-code 110059.SH]
  ashare ingest cb-holders -start 20240101 -end 20251231 [-code 110059.SH]

说明:
  全市场按交易日拉截面，请先 ingest calendar。
  -code 可写 000001.SZ 或 000001（需已有证券列表）。
  cb-daily：-date 拉某日全市场截面；或 -start/-end，可加 -code。
  cb-holders：-period / -start/-end / -code；-code 省略时同步库内全部转债。
  cb-basic、cb-daily 约 2000 积分；cb-holders 约 5000 积分。
  限流见 config.yaml 的 tushare.rpm.default / tushare.rpm.apis（未配置的接口用 default）。
  也可在 Web「采集任务」页触发同步。
`)
}
