package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
	"github.com/stock/ashare/internal/research"
)

func cmdShow(ctx context.Context, cfg config.Config, args []string) error {
	code, start, end, qfq, err := parseShowFlags(args)
	if err != nil {
		return err
	}
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	instruments := repo.NewInstrumentRepo(gdb)
	market := repo.NewMarketRepo(gdb)
	tsCode, err := ingest.ResolveTSCode(ctx, instruments, code)
	if err != nil {
		return err
	}
	inst, _ := instruments.GetByTSCode(ctx, tsCode)

	bars, err := market.ListDailyBars(ctx, tsCode, start, end, 60, qfq)
	if err != nil {
		return err
	}
	if len(bars) == 0 {
		return fmt.Errorf("%s 没有日线，请先 ingest daily", tsCode)
	}

	title := tsCode
	if inst.Name != "" {
		title = fmt.Sprintf("%s %s", tsCode, inst.Name)
	}
	kind := "未复权"
	if qfq {
		kind = "前复权"
	}
	fmt.Printf("%s  (%s)  %s ~ %s  %d根\n",
		title, kind, dates.ISO(bars[0].TradeDate), dates.ISO(bars[len(bars)-1].TradeDate), len(bars))
	fmt.Printf("%-10s %8s %8s %8s %8s %12s\n", "date", "open", "high", "low", "close", "volume")
	for _, b := range bars {
		fmt.Printf(" %-10s %8.2f %8.2f %8.2f %8.2f %12d\n",
			dates.ISO(b.TradeDate), b.Open, b.High, b.Low, b.Close, b.Volume)
	}
	return nil
}

func cmdStats(ctx context.Context, cfg config.Config, args []string) error {
	code, start, end, _, err := parseShowFlags(args)
	if err != nil {
		return err
	}
	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	instruments := repo.NewInstrumentRepo(gdb)
	market := repo.NewMarketRepo(gdb)
	tsCode, err := ingest.ResolveTSCode(ctx, instruments, code)
	if err != nil {
		return err
	}
	inst, _ := instruments.GetByTSCode(ctx, tsCode)

	bars, err := market.ListDailyBars(ctx, tsCode, start, end, 0, true)
	if err != nil {
		return err
	}
	if len(bars) == 0 {
		return fmt.Errorf("%s 没有前复权日线（需同时有 daily_bar 和 adj_factor）", tsCode)
	}
	st, err := research.FromBars(bars)
	if err != nil {
		return err
	}
	name := tsCode
	if inst.Name != "" {
		name = tsCode + " " + inst.Name
	}
	fmt.Printf("%s  前复权收益\n", name)
	fmt.Printf("区间     %s ~ %s  (%d 根)\n", dates.ISO(st.Start), dates.ISO(st.End), st.N)
	fmt.Printf("收盘     %.4f -> %.4f\n", st.StartClose, st.EndClose)
	fmt.Printf("累计收益 %+8.2f%%\n", st.TotalReturn*100)
	fmt.Printf("年化     %+8.2f%%\n", st.AnnReturn*100)
	fmt.Printf("最大回撤 %+8.2f%%\n", st.MaxDrawdown*100)
	return nil
}

func parseShowFlags(args []string) (code string, start, end time.Time, qfq bool, err error) {
	if len(args) == 0 {
		return "", time.Time{}, time.Time{}, false, fmt.Errorf("请提供证券代码，例如 000001.SZ")
	}
	code = args[0]
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	startStr := fs.String("start", "", "开始日期")
	endStr := fs.String("end", "", "结束日期")
	qfqFlag := fs.Bool("qfq", false, "使用前复权")
	if err := fs.Parse(args[1:]); err != nil {
		return "", time.Time{}, time.Time{}, false, err
	}
	if strings.TrimSpace(*startStr) != "" {
		start, err = dates.Parse(*startStr)
		if err != nil {
			return "", time.Time{}, time.Time{}, false, err
		}
	}
	if strings.TrimSpace(*endStr) != "" {
		end, err = dates.Parse(*endStr)
		if err != nil {
			return "", time.Time{}, time.Time{}, false, err
		}
	}
	return code, start, end, *qfqFlag, nil
}
