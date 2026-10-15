package backtest

import "time"

type Params struct {
	Code           string
	Fast           int
	Slow           int
	InitCash       float64
	CommissionRate float64
	CommissionMin  float64
	StampTaxRate   float64
	Slippage       float64
	LotSize        int64
	TradeFrom      time.Time
	TradeTo        time.Time
}

func DefaultParams() Params {
	return Params{
		Fast:           5,
		Slow:           20,
		InitCash:       100_000,
		CommissionRate: 0.00025,
		CommissionMin:  5,
		StampTaxRate:   0.0005,
		Slippage:       0.001,
		LotSize:        100,
	}
}

func (p Params) withDefaults() Params {
	d := DefaultParams()
	if p.Fast <= 0 {
		p.Fast = d.Fast
	}
	if p.Slow <= 0 {
		p.Slow = d.Slow
	}
	if p.InitCash <= 0 {
		p.InitCash = d.InitCash
	}
	if p.CommissionRate < 0 {
		p.CommissionRate = d.CommissionRate
	}
	if p.CommissionMin < 0 {
		p.CommissionMin = d.CommissionMin
	}
	if p.StampTaxRate < 0 {
		p.StampTaxRate = d.StampTaxRate
	}
	if p.Slippage < 0 {
		p.Slippage = d.Slippage
	}
	if p.LotSize <= 0 {
		p.LotSize = d.LotSize
	}
	return p
}

func KnownLimitations() []string {
	return []string{
		"未模拟现金分红到账和再投资",
		"未模拟退市整理、摘牌后强制了结",
		"ST 仅有当前名称标记，没有历史 ST 快照",
		"买卖单位统一 100 股，未区分科创板 200 股等规则",
		"假设开盘价可全额成交，未限制成交量",
		"未模拟集合竞价、盘中滑点路径，滑点按开盘价固定比例",
		"涨跌停若无标记，仅在有涨跌停价时用开盘价比较",
	}
}
