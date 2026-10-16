package domain

import "time"

// IndexBasic 指数基本信息。自然键：TSCode。对齐 Tushare index_basic。
type IndexBasic struct {
	Audit
	TSCode      string
	Name        string
	Fullname    string
	Market      string
	Publisher   string
	IndexType   string
	Category    string
	BaseDate    *time.Time
	BasePoint   *float64
	ListDate    *time.Time
	WeightRule  string
	Description string `gorm:"column:description"`
	ExpDate     *time.Time
}

func (IndexBasic) TableName() string { return "index_basic" }

// IndexDaily 指数日线。自然键：TSCode + TradeDate。对齐 Tushare index_daily。
type IndexDaily struct {
	Audit
	TSCode    string
	TradeDate time.Time
	Close     *float64
	Open      *float64
	High      *float64
	Low       *float64
	PreClose  *float64
	Change    *float64
	PctChg    *float64
	Vol       *float64 // 手
	Amount    *float64 // 千元
}

func (IndexDaily) TableName() string { return "index_daily" }

// IndexClassify 申万行业分类。自然键：IndexCode + Src。对齐 Tushare index_classify。
type IndexClassify struct {
	Audit
	IndexCode    string
	IndustryName string
	ParentCode   string
	Level        string
	IndustryCode string
	IsPub        string
	Src          string
}

func (IndexClassify) TableName() string { return "index_classify" }

// SWDaily 申万行业日线。自然键：TSCode + TradeDate。对齐 Tushare sw_daily。
type SWDaily struct {
	Audit
	TSCode    string
	TradeDate time.Time
	Name      string
	Open      *float64
	Low       *float64
	High      *float64
	Close     *float64
	Change    *float64
	PctChange *float64
	Vol       *float64 // 万股
	Amount    *float64 // 万元
	PE        *float64 `gorm:"column:pe"`
	PB        *float64 `gorm:"column:pb"`
	FloatMV   *float64 `gorm:"column:float_mv"`
	TotalMV   *float64 `gorm:"column:total_mv"`
}

func (SWDaily) TableName() string { return "sw_daily" }
