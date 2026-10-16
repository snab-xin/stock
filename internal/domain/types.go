package domain

import "time"

// Audit 是持久化实体的公共字段。业务唯一性仍由各表的自然键保证。
type Audit struct {
	ID        int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Instrument 证券主数据。自然键：TSCode。股票字段对齐 Tushare stock_basic。
type Instrument struct {
	Audit
	TSCode     string
	Symbol     string
	Name       string
	Area       string
	Industry   string
	Fullname   string
	Enname     string
	Cnspell    string
	Market     string
	Exchange   string
	CurrType   string
	ListStatus string `gorm:"column:list_status"`
	ListDate   *time.Time
	DelistDate *time.Time
	IsHS       string `gorm:"column:is_hs"`
	ActName    string
	ActEntType string
	AssetType  string
	IsST       bool `gorm:"column:is_st"`
}

func (Instrument) TableName() string { return "instrument" }

// TradeCalendar 交易所交易日历。自然键：Exchange + CalDate。
type TradeCalendar struct {
	Audit
	Exchange     string
	CalDate      time.Time
	IsOpen       bool
	PretradeDate *time.Time
}

func (TradeCalendar) TableName() string { return "trade_calendar" }

// DailyBar 未复权日线。自然键：TSCode + TradeDate。
// 行情字段对齐 Tushare daily；涨跌停来自 stk_limit，换手率来自 daily_basic。
type DailyBar struct {
	Audit
	TSCode      string
	TradeDate   time.Time
	Open        float64
	High        float64
	Low         float64
	Close       float64
	PreClose    *float64
	Change      *float64
	PctChg      *float64
	Volume      int64   // 股
	Amount      float64 // 元
	AHVolume    int64   `gorm:"column:ah_vol"`    // 盘后成交量（股）
	AHAmount    float64 `gorm:"column:ah_amount"` // 盘后成交额（元）
	Turnover    *float64
	HighLimit   *float64
	LowLimit    *float64
	IsLimitUp   *bool
	IsLimitDown *bool
	IsPaused    bool
}

func (DailyBar) TableName() string { return "daily_bar" }

// AdjFactor 复权因子。自然键：TSCode + TradeDate。
type AdjFactor struct {
	Audit
	TSCode    string
	TradeDate time.Time
	Factor    float64 `gorm:"column:adj_factor"`
}

func (AdjFactor) TableName() string { return "adj_factor" }

// Suspension 停复牌。自然键：TSCode + TradeDate。对齐 Tushare suspend_d。
type Suspension struct {
	Audit
	TSCode        string
	TradeDate     time.Time
	SuspendTiming string
	SuspendType   string
}

func (Suspension) TableName() string { return "suspension" }

// StkLimit 每日涨跌停价格。自然键：TSCode + TradeDate。对齐 Tushare stk_limit。
type StkLimit struct {
	Audit
	TradeDate time.Time
	TSCode    string
	PreClose  *float64
	UpLimit   *float64
	DownLimit *float64
	AssetType string
	Exchange  string
}

func (StkLimit) TableName() string { return "stk_limit" }

// DailyBasic 每日指标。自然键：TSCode + TradeDate。对齐 Tushare daily_basic。
type DailyBasic struct {
	Audit
	TSCode        string
	TradeDate     time.Time
	Close         *float64
	TurnoverRate  *float64
	TurnoverRateF *float64 `gorm:"column:turnover_rate_f"`
	VolumeRatio   *float64
	PE            *float64 `gorm:"column:pe"`
	PETTM         *float64 `gorm:"column:pe_ttm"`
	PB            *float64 `gorm:"column:pb"`
	PS            *float64 `gorm:"column:ps"`
	PSTTM         *float64 `gorm:"column:ps_ttm"`
	DVRatio       *float64 `gorm:"column:dv_ratio"`
	DVTTM         *float64 `gorm:"column:dv_ttm"`
	TotalShare    *float64
	FloatShare    *float64
	FreeShare     *float64
	TotalMV       *float64 `gorm:"column:total_mv"`
	CircMV        *float64 `gorm:"column:circ_mv"`
	LimitStatus   *int
}

func (DailyBasic) TableName() string { return "daily_basic" }

// IngestRun 采集任务流水。主键即 ID。
type IngestRun struct {
	Audit
	JobName      string
	StartedAt    time.Time
	FinishedAt   *time.Time
	Status       string
	Message      string
	RowsAffected int64
	Params       JSONObject `gorm:"type:jsonb"`
}

func (IngestRun) TableName() string { return "ingest_run" }

// CBBasic 可转债基础信息。自然键：TSCode。来源 Tushare cb_basic。
type CBBasic struct {
	Audit
	TSCode            string
	BondFullName      string
	BondShortName     string
	CBCode            string `gorm:"column:cb_code"`
	CBType            string `gorm:"column:cb_type"`
	StkCode           string
	StkShortName      string
	Maturity          *float64
	Par               *float64
	IssuePrice        *float64
	IssueSize         *float64
	RemainSize        *float64
	ValueDate         *time.Time
	MaturityDate      *time.Time
	RateType          string
	CouponRate        *float64
	AddRate           *float64
	PayPerYear        *int
	ListDate          *time.Time
	DelistDate        *time.Time
	Exchange          string
	ConvStartDate     *time.Time
	ConvEndDate       *time.Time
	ConvStopDate      *time.Time
	FirstConvPrice    *float64
	ConvPrice         *float64
	RateClause        string
	PutClause         string
	MaturityCallPrice *float64
	CallClause        string
	ResetClause       string
	ConvClause        string
	Guarantor         string
	GuaranteeType     string
	IssueRating       string
	NewestRating      string
	RatingComp        string
}

func (CBBasic) TableName() string { return "cb_basic" }

// CBTopHolder 可转债前十大持有人。自然键：TSCode + EndDate + HolderRank。
type CBTopHolder struct {
	Audit
	TSCode     string
	EndDate    time.Time
	HolderRank int
	HolderName string
	HoldAmount *float64 // 万张
	HoldRatio  *float64 // %
	Type       int      // 0:自动抓取，1:手动更新
}

func (CBTopHolder) TableName() string { return "cb_top_holder" }

// CBDaily 可转债日线。自然键：TSCode + TradeDate。来源 Tushare cb_daily。
type CBDaily struct {
	Audit
	TSCode       string
	TradeDate    time.Time
	PreClose     *float64
	Open         *float64
	High         *float64
	Low          *float64
	Close        *float64
	Change       *float64
	PctChg       *float64
	Volume       *float64 // 手
	Amount       *float64 // 万元
	BondValue    *float64
	BondOverRate *float64
	CBValue      *float64 `gorm:"column:cb_value"`
	CBOverRate   *float64 `gorm:"column:cb_over_rate"`
}

func (CBDaily) TableName() string { return "cb_daily" }
