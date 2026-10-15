package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/backtest"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/internal/repo"
)

func cmdBacktest(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请提供证券代码，例如 000001.SZ")
	}
	code := args[0]
	fs := flag.NewFlagSet("backtest", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	startStr := fs.String("start", "", "开始日期（信号预热会再往前取）")
	endStr := fs.String("end", "", "结束日期")
	fast := fs.Int("fast", cfg.Backtest.Fast, "快均线周期")
	slow := fs.Int("slow", cfg.Backtest.Slow, "慢均线周期")
	cash := fs.Float64("cash", cfg.Backtest.InitCash, "初始资金")
	out := fs.String("out", "", "把完整结果写成 JSON 文件")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	p := paramsFromConfig(cfg)
	p.Fast = *fast
	p.Slow = *slow
	p.InitCash = *cash
	if strings.TrimSpace(*startStr) != "" {
		t, err := dates.Parse(*startStr)
		if err != nil {
			return err
		}
		p.TradeFrom = t
	}
	if strings.TrimSpace(*endStr) != "" {
		t, err := dates.Parse(*endStr)
		if err != nil {
			return err
		}
		p.TradeTo = t
	}

	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	res, err := backtest.LoadAndRun(ctx, repo.NewInstrumentRepo(gdb), repo.NewMarketRepo(gdb), code, p)
	if err != nil {
		return err
	}
	printBacktest(res)

	if strings.TrimSpace(*out) != "" {
		if err := writeBacktestJSON(*out, res); err != nil {
			return err
		}
		fmt.Printf("完整结果已写入 %s\n", *out)
	}
	return nil
}

func paramsFromConfig(cfg config.Config) backtest.Params {
	b := cfg.Backtest
	return backtest.Params{
		Fast:           b.Fast,
		Slow:           b.Slow,
		InitCash:       b.InitCash,
		CommissionRate: b.CommissionRate,
		CommissionMin:  b.CommissionMin,
		StampTaxRate:   b.StampTaxRate,
		Slippage:       b.Slippage,
		LotSize:        b.LotSize,
	}
}

func printBacktest(res backtest.Result) {
	title := res.Code
	if res.Name != "" {
		title = res.Code + " " + res.Name
	}
	fmt.Printf("%s  双均线 %d/%d\n", title, res.Fast, res.Slow)
	fmt.Printf("区间     %s ~ %s  （对齐 %d 根）\n", res.Start, res.End, res.BarCount)
	fmt.Printf("资金     %.2f -> %.2f\n", res.InitCash, res.EndEquity)
	fmt.Printf("策略收益 %+8.2f%%  年化 %+8.2f%%  最大回撤 %+8.2f%%\n",
		res.TotalReturn*100, res.AnnReturn*100, res.MaxDrawdown*100)
	fmt.Printf("买入持有 %+8.2f%%  （同期前复权收盘）\n", res.BuyHoldReturn*100)
	fmt.Printf("成交     %d 笔  完成回合胜率 %.0f%%\n", res.TradeCount, res.WinRate*100)
	fmt.Printf("费用     佣金 %.2f  印花税 %.2f  滑点 %.2f\n", res.Commission, res.StampTax, res.SlippageCost)
	fmt.Println()
	fmt.Printf("%-10s %-6s %8s %8s %10s %8s %8s  %s\n",
		"date", "side", "price", "shares", "amount", "comm", "stamp", "reason")
	n := 0
	for _, t := range res.Trades {
		if t.Side != "buy" && t.Side != "sell" {
			continue
		}
		fmt.Printf(" %-10s %-6s %8.2f %8d %10.2f %8.2f %8.2f  %s\n",
			dates.ISO(t.Date), t.Side, t.Price, t.Shares, t.Amount, t.Commission, t.StampTax, t.Reason)
		n++
	}
	if n == 0 {
		fmt.Println(" (无成交)")
	}
	fmt.Println()
	fmt.Println("未覆盖规则:")
	for _, line := range res.Limitations {
		fmt.Printf("  - %s\n", line)
	}
}

func writeBacktestJSON(path string, res backtest.Result) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
