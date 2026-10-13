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
