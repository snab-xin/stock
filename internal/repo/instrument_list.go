package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
)

// StockListRow 列表用：证券主数据 + 最新（或指定日）日线行情。
type StockListRow struct {
	domain.Instrument
	QuoteDate   *time.Time `gorm:"column:quote_date"`
	PreClose    *float64   `gorm:"column:pre_close"`
	Close       *float64   `gorm:"column:close"`
	Change      *float64   `gorm:"column:change"`
	PctChg      *float64   `gorm:"column:pct_chg"`
	Volume      *int64     `gorm:"column:volume"`
	Amount      *float64   `gorm:"column:amount"`
	IsLimitUp   *bool      `gorm:"column:is_limit_up"`
	IsLimitDown *bool      `gorm:"column:is_limit_down"`
	IsPaused    *bool      `gorm:"column:is_paused"`
}

// StockListQuery 股票列表检索。SortBy 见 stockSortColumns；SortDesc 为 true 时降序。
type StockListQuery struct {
	Q          string
	Limit      int
	Offset     int
	ListedOnly bool
	TradeDate  time.Time
	SortBy     string
	SortDesc   bool
}

var stockSortColumns = map[string]string{
	"ts_code":    "i.ts_code",
	"name":       "i.name",
	"close":      "q.close",
	"change":     "q.change",
	"pct_chg":    "q.pct_chg",
	"volume":     "q.volume",
	"amount":     "q.amount",
	"quote_date": "q.quote_date",
	"list_date":  "i.list_date",
}

func stockOrderBy(sortBy string, sortDesc bool) string {
	col, ok := stockSortColumns[strings.TrimSpace(sortBy)]
	if !ok {
		return "q.pct_chg DESC NULLS LAST, i.list_date DESC NULLS LAST, i.ts_code"
	}
	dir := "ASC"
	if sortDesc {
		dir = "DESC"
	}
	return col + " " + dir + " NULLS LAST, i.ts_code ASC"
}

func stockListJoinWhere(query StockListQuery) (joinSQL, where string, args []any) {
	q := strings.TrimSpace(query.Q)
	joinSQL = `
LEFT JOIN LATERAL (
  SELECT d.trade_date AS quote_date,
         d.pre_close,
         d.close,
         CASE WHEN d.pre_close IS NULL THEN NULL ELSE d.close - d.pre_close END AS change,
         CASE
           WHEN d.pre_close IS NULL OR d.pre_close = 0 THEN NULL
           ELSE (d.close - d.pre_close) / d.pre_close * 100
         END AS pct_chg,
         d.volume,
         d.amount,
         d.is_limit_up,
         d.is_limit_down,
         d.is_paused
  FROM daily_bar d
  WHERE d.ts_code = i.ts_code`
	if !query.TradeDate.IsZero() {
		joinSQL += ` AND d.trade_date = ?`
		args = append(args, query.TradeDate)
	}
	joinSQL += `
  ORDER BY d.trade_date DESC
  LIMIT 1
) q ON true`

	where = "i.asset_type = 'stock'"
	if q != "" {
		like := "%" + q + "%"
		where += " AND (i.ts_code ILIKE ? OR i.symbol ILIKE ? OR i.name ILIKE ?)"
		args = append(args, like, like, like)
	}
	if query.ListedOnly {
		where += " AND i.status = 'L'"
	}
	return joinSQL, where, args
}

// SearchWithQuote 股票列表：主数据 + 最新日线。
func (r *InstrumentRepo) SearchWithQuote(ctx context.Context, query StockListQuery) ([]StockListRow, int64, error) {
	limit, offset := query.Limit, query.Offset
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	countDB := r.db.WithContext(ctx).Model(&domain.Instrument{}).Where("asset_type = ?", "stock")
	q := strings.TrimSpace(query.Q)
	if q != "" {
		like := "%" + q + "%"
		countDB = countDB.Where("ts_code ILIKE ? OR symbol ILIKE ? OR name ILIKE ?", like, like, like)
	}
	if query.ListedOnly {
		countDB = countDB.Where("status = ?", "L")
	}
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	joinSQL, where, args := stockListJoinWhere(query)
	sql := `
SELECT i.*, q.quote_date, q.pre_close, q.close, q.change, q.pct_chg,
       q.volume, q.amount, q.is_limit_up, q.is_limit_down, q.is_paused
FROM instrument i
` + joinSQL + `
WHERE ` + where + `
ORDER BY ` + stockOrderBy(query.SortBy, query.SortDesc) + `
LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var rows []StockListRow
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// StockNeighbors 当前股票在相同筛选/排序下的相邻代码。Index 从 1 起；不在结果集时 Index=0。
type StockNeighbors struct {
	Prev  string
	Next  string
	Index int64
	Total int64
}

func (r *InstrumentRepo) Neighbors(ctx context.Context, tsCode string, query StockListQuery) (StockNeighbors, error) {
	tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
	if tsCode == "" {
		return StockNeighbors{}, fmt.Errorf("缺少股票代码")
	}
	joinSQL, where, args := stockListJoinWhere(query)
	orderBy := stockOrderBy(query.SortBy, query.SortDesc)
	sql := `
WITH ranked AS (
  SELECT i.ts_code,
         LAG(i.ts_code) OVER (ORDER BY ` + orderBy + `) AS prev_code,
         LEAD(i.ts_code) OVER (ORDER BY ` + orderBy + `) AS next_code,
         ROW_NUMBER() OVER (ORDER BY ` + orderBy + `) AS rn,
         COUNT(*) OVER () AS total
  FROM instrument i
` + joinSQL + `
  WHERE ` + where + `
)
SELECT prev_code, next_code, rn, total FROM ranked WHERE ts_code = ?`
	args = append(args, tsCode)

	var row struct {
		PrevCode *string `gorm:"column:prev_code"`
		NextCode *string `gorm:"column:next_code"`
		RN       int64   `gorm:"column:rn"`
		Total    int64   `gorm:"column:total"`
	}
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&row).Error; err != nil {
		return StockNeighbors{}, err
	}
	out := StockNeighbors{Index: row.RN, Total: row.Total}
	if row.PrevCode != nil {
		out.Prev = *row.PrevCode
	}
	if row.NextCode != nil {
		out.Next = *row.NextCode
	}
	return out, nil
}
