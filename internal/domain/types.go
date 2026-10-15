package domain

import "time"

// Audit 是持久化实体的公共字段。业务唯一性仍由各表的自然键保证。
type Audit struct {
	ID        int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Instrument 证券主数据。自然键：TSCode。
type Instrument struct {
	Audit
	TSCode     string
	Symbol     string
	Name       string
	Exchange   string
	AssetType  string
	ListDate   *time.Time
	DelistDate *time.Time
	IsST       bool `gorm:"column:is_st"`
	Status     string
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
type DailyBar struct {
	Audit
	TSCode      string
	TradeDate   time.Time
	Open        float64
	High        float64
	Low         float64
	Close       float64
	PreClose    *float64
	Volume      int64
	Amount      float64
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

// Suspension 停牌。自然键：TSCode + TradeDate。
type Suspension struct {
	Audit
	TSCode      string
	TradeDate   time.Time
	SuspendType string
	Reason      string
}

func (Suspension) TableName() string { return "suspension" }

// IndexConstituent 指数成分快照。自然键：IndexCode + TSCode + TradeDate。
type IndexConstituent struct {
	Audit
	IndexCode string
	TSCode    string
	TradeDate time.Time
	Weight    *float64
}

func (IndexConstituent) TableName() string { return "index_constituent" }

// IngestRun 采集任务流水。主键即 ID。
type IngestRun struct {
	Audit
	JobName      string
	StartedAt    time.Time
	FinishedAt   *time.Time
	Status       string
	Message      string
	RowsAffected int64
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
