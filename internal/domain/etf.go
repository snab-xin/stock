package domain

import "time"

// ETFBasic ETF基础信息。自然键：TSCode。对齐 Tushare etf_basic。
type ETFBasic struct {
	Audit
	TSCode     string
	Csname     string
	Extname    string
	Cname      string
	IndexCode  string
	IndexName  string
	SetupDate  *time.Time
	ListDate   *time.Time
	ListStatus string
	Exchange   string
	MgrName    string
	CustodName string
	MgtFee     *float64
	ETFType    string `gorm:"column:etf_type"`
}

func (ETFBasic) TableName() string { return "etf_basic" }

// ETFDaily ETF日线。自然键：TSCode + TradeDate。对齐 Tushare fund_daily。
type ETFDaily struct {
	Audit
	TSCode    string
	TradeDate time.Time
	Open      *float64
	High      *float64
	Low       *float64
	Close     *float64
	PreClose  *float64
	Change    *float64
	PctChg    *float64
	Vol       *float64 // 手
	Amount    *float64 // 千元
}

func (ETFDaily) TableName() string { return "etf_daily" }

// ETFAdj ETF复权因子。自然键：TSCode + TradeDate。对齐 Tushare fund_adj。
type ETFAdj struct {
	Audit
	TSCode    string
	TradeDate time.Time
	Factor    float64 `gorm:"column:adj_factor"`
}

func (ETFAdj) TableName() string { return "etf_adj" }

// ETFShareSize ETF份额规模。自然键：TSCode + TradeDate。对齐 Tushare etf_share_size。
type ETFShareSize struct {
	Audit
	TradeDate  time.Time
	TSCode     string
	ETFName    string `gorm:"column:etf_name"`
	TotalShare *float64
	TotalSize  *float64
	NAV        *float64 `gorm:"column:nav"`
	Close      *float64
	Exchange   string
}

func (ETFShareSize) TableName() string { return "etf_share_size" }

// ETFSHCons 沪市 ETF 每日持仓组合。自然键：TSCode + TradeDate + ConCode。对齐 Tushare etf_sh_cons。
type ETFSHCons struct {
	Audit
	TradeDate time.Time
	TSCode    string
	ConCode   string
	ConName   string
	Qty       *int64
	SubFlag   string
	CPR       *float64 `gorm:"column:cpr"`
	RDR       *float64 `gorm:"column:rdr"`
	SCA       *float64 `gorm:"column:sca"`
	Exchange  string
}

func (ETFSHCons) TableName() string { return "etf_sh_cons" }

// ETFSZCons 深市 ETF 每日持仓组合。自然键：TSCode + TradeDate + ConCode。对齐 Tushare etf_sz_cons。
type ETFSZCons struct {
	Audit
	TradeDate time.Time
	TSCode    string
	ConCode   string
	ConName   string
	Qty       *int64
	SubFlag   string
	CPR       *float64 `gorm:"column:cpr"`
	RDR       *float64 `gorm:"column:rdr"`
	SubCC     *float64 `gorm:"column:sub_cc"`
	RedCC     *float64 `gorm:"column:red_cc"`
	Exchange  string
}

func (ETFSZCons) TableName() string { return "etf_sz_cons" }
