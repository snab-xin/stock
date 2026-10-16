package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type CBRepo struct {
	db *gorm.DB
}

func NewCBRepo(db *gorm.DB) *CBRepo {
	return &CBRepo{db: db}
}

type CBHolderSummary struct {
	EndDate     time.Time
	Top10Ratio  float64
	Top10Amount float64
	HolderCount int64
}

func (r *CBRepo) UpsertBasic(ctx context.Context, items []domain.CBBasic) error {
	return upsert(ctx, r.db, &items, []string{"ts_code"}, []string{
		"bond_full_name", "bond_short_name", "cb_code", "cb_type",
		"stk_code", "stk_short_name", "maturity", "par", "issue_price",
		"issue_size", "remain_size", "value_date", "maturity_date",
		"rate_type", "coupon_rate", "add_rate", "pay_per_year",
		"list_date", "delist_date", "exchange",
		"conv_start_date", "conv_end_date", "conv_stop_date",
		"first_conv_price", "conv_price",
		"rate_clause", "put_clause", "maturity_call_price",
		"call_clause", "reset_clause", "conv_clause",
		"guarantor", "guarantee_type", "issue_rating", "newest_rating", "rating_comp",
	})
}

func (r *CBRepo) UpsertTopHolders(ctx context.Context, items []domain.CBTopHolder) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "end_date", "holder_rank"}, []string{
		"holder_name", "hold_amount", "hold_ratio",
	})
}

// ReplaceTopHolders 用新名单整期覆盖：先删该转债该报告期全部行，再写入。
func (r *CBRepo) ReplaceTopHolders(ctx context.Context, tsCode string, endDate time.Time, items []domain.CBTopHolder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ts_code = ? AND end_date = ?", tsCode, endDate).
			Delete(&domain.CBTopHolder{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return upsert(ctx, tx, &items, []string{"ts_code", "end_date", "holder_rank"}, []string{
			"holder_name", "hold_amount", "hold_ratio",
		})
	})
}

func (r *CBRepo) ListTSCodes(ctx context.Context) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Model(&domain.CBBasic{}).
		Order("ts_code").
		Pluck("ts_code", &codes).Error
	return codes, err
}

func (r *CBRepo) GetByTSCode(ctx context.Context, tsCode string) (domain.CBBasic, error) {
	var items []domain.CBBasic
	if err := r.db.WithContext(ctx).Where("ts_code = ?", tsCode).Limit(1).Find(&items).Error; err != nil {
		return domain.CBBasic{}, err
	}
	if len(items) == 0 {
		return domain.CBBasic{}, gorm.ErrRecordNotFound
	}
	return items[0], nil
}

func (r *CBRepo) Search(ctx context.Context, q string, limit, offset int, listedOnly bool) ([]domain.CBBasic, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	db := r.db.WithContext(ctx).Model(&domain.CBBasic{})
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		db = db.Where(
			"ts_code ILIKE ? OR bond_short_name ILIKE ? OR bond_full_name ILIKE ? OR stk_code ILIKE ? OR stk_short_name ILIKE ?",
			like, like, like, like, like,
		)
	}
	if listedOnly {
		db = db.Where("delist_date IS NULL")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.CBBasic
	err := db.Order("list_date DESC NULLS LAST, ts_code").
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	return items, total, err
}

func (r *CBRepo) ListHolderPeriods(ctx context.Context, tsCode string) ([]time.Time, error) {
	var dates []time.Time
	err := r.db.WithContext(ctx).
		Raw(`SELECT DISTINCT end_date FROM cb_top_holder WHERE ts_code = ? ORDER BY end_date DESC`, tsCode).
		Scan(&dates).Error
	return dates, err
}

func (r *CBRepo) ListTopHolders(ctx context.Context, tsCode string, endDate time.Time) ([]domain.CBTopHolder, error) {
	var items []domain.CBTopHolder
	err := r.db.WithContext(ctx).
		Where("ts_code = ? AND end_date = ?", tsCode, endDate).
		Order("holder_rank").
		Find(&items).Error
	return items, err
}

func (r *CBRepo) UpsertDaily(ctx context.Context, items []domain.CBDaily) error {
	return upsert(ctx, r.db, &items, []string{"ts_code", "trade_date"}, []string{
		"pre_close", "open", "high", "low", "close", "change", "pct_chg",
		"volume", "amount", "bond_value", "bond_over_rate", "cb_value", "cb_over_rate",
	})
}

// CBListRow 列表用：基础信息 + 最新（或指定日）行情 + 流通规模。
type CBListRow struct {
	domain.CBBasic
	QuoteDate   *time.Time `gorm:"column:quote_date"`
	PreClose    *float64   `gorm:"column:pre_close"`
	Close       *float64   `gorm:"column:close"`
	Change      *float64   `gorm:"column:change"`
	PctChg      *float64   `gorm:"column:pct_chg"`
	Volume      *float64   `gorm:"column:volume"`
	Amount      *float64   `gorm:"column:amount"`
	CBOverRate  *float64   `gorm:"column:cb_over_rate"` // 转股溢价率 %
	Top10Ratio  *float64   `gorm:"column:top10_ratio"`  // 最近一期十大合计持仓占比 %
	HolderCount *int64     `gorm:"column:holder_count"`
	FloatSize   *float64   `gorm:"column:float_size"` // 流通规模（元）= 余额 × (1 - 十大占比)
}

// CBListQuery 可转债列表检索。SortBy 见 cbSortColumns；SortDesc 为 true 时降序。
type CBListQuery struct {
	Q          string
	Limit      int
	Offset     int
	ListedOnly bool
	TradeDate  time.Time
	SortBy     string
	SortDesc   bool
}

var cbSortColumns = map[string]string{
	"ts_code":             "b.ts_code",
	"bond_short_name":     "b.bond_short_name",
	"close":               "q.close",
	"change":              "q.change",
	"pct_chg":             "q.pct_chg",
	"cb_over_rate":        "q.cb_over_rate",
	"stk":                 "b.stk_code",
	"conv_price":          "b.conv_price",
	"maturity_call_price": "b.maturity_call_price",
	"rating":              "COALESCE(NULLIF(b.newest_rating, ''), b.issue_rating)",
	"quote_date":          "q.quote_date",
	"issue_size":          "b.issue_size",
	"remain_size":         "b.remain_size",
	"float_size":          "(" + cbFloatSizeExpr + ")",
}

func cbOrderBy(sortBy string, sortDesc bool) string {
	col, ok := cbSortColumns[strings.TrimSpace(sortBy)]
	if !ok {
		return "q.pct_chg DESC NULLS LAST, b.list_date DESC NULLS LAST, b.ts_code"
	}
	dir := "ASC"
	if sortDesc {
		dir = "DESC"
	}
	return col + " " + dir + " NULLS LAST, b.ts_code ASC"
}

// cbFloatSizeExpr 流通规模：无十大持有人时等于余额，否则余额 × (1 - 十大占比)。
const cbFloatSizeExpr = `CASE WHEN b.remain_size IS NULL THEN NULL WHEN th.holder_count IS NULL OR th.holder_count = 0 THEN b.remain_size ELSE b.remain_size * (100.0 - COALESCE(th.top10_ratio, 0)) / 100.0 END`

// cbListJoinWhere 拼列表用的行情 / 十大持仓 lateral join 与筛选条件（不含 ORDER/LIMIT）。
func cbListJoinWhere(query CBListQuery) (joinSQL, where string, args []any) {
	q := strings.TrimSpace(query.Q)
	joinSQL = `
LEFT JOIN LATERAL (
  SELECT d.trade_date AS quote_date, d.pre_close, d.close, d.change, d.pct_chg,
         d.volume, d.amount, d.cb_over_rate
  FROM cb_daily d
  WHERE d.ts_code = b.ts_code`
	if !query.TradeDate.IsZero() {
		joinSQL += ` AND d.trade_date = ?`
		args = append(args, query.TradeDate)
	}
	joinSQL += `
  ORDER BY d.trade_date DESC
  LIMIT 1
) q ON true
LEFT JOIN LATERAL (
  SELECT COALESCE(SUM(h.hold_ratio), 0) AS top10_ratio,
         COUNT(*)::bigint AS holder_count
  FROM cb_top_holder h
  WHERE h.ts_code = b.ts_code
    AND h.end_date = (
      SELECT MAX(x.end_date) FROM cb_top_holder x WHERE x.ts_code = b.ts_code
    )
) th ON true`

	where = "1=1"
	if q != "" {
		like := "%" + q + "%"
		where += " AND (b.ts_code ILIKE ? OR b.bond_short_name ILIKE ? OR b.bond_full_name ILIKE ? OR b.stk_code ILIKE ? OR b.stk_short_name ILIKE ?)"
		args = append(args, like, like, like, like, like)
	}
	if query.ListedOnly {
		where += " AND b.delist_date IS NULL"
	}
	return joinSQL, where, args
}

func (r *CBRepo) SearchWithQuote(ctx context.Context, query CBListQuery) ([]CBListRow, int64, error) {
	limit, offset := query.Limit, query.Offset
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	countDB := r.db.WithContext(ctx).Model(&domain.CBBasic{})
	q := strings.TrimSpace(query.Q)
	if q != "" {
		like := "%" + q + "%"
		countDB = countDB.Where(
			"ts_code ILIKE ? OR bond_short_name ILIKE ? OR bond_full_name ILIKE ? OR stk_code ILIKE ? OR stk_short_name ILIKE ?",
			like, like, like, like, like,
		)
	}
	if query.ListedOnly {
		countDB = countDB.Where("delist_date IS NULL")
	}
	var total int64
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	joinSQL, where, args := cbListJoinWhere(query)
	sql := `
SELECT b.*, q.quote_date, q.pre_close, q.close, q.change, q.pct_chg,
       q.volume, q.amount, q.cb_over_rate,
       th.top10_ratio, th.holder_count,
       ` + cbFloatSizeExpr + ` AS float_size
FROM cb_basic b
` + joinSQL + `
WHERE ` + where + `
ORDER BY ` + cbOrderBy(query.SortBy, query.SortDesc) + `
LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var rows []CBListRow
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// CBNeighbors 当前转债在相同筛选/排序下的相邻代码。Index 从 1 起；不在结果集时 Index=0。
type CBNeighbors struct {
	Prev  string
	Next  string
	Index int64
	Total int64
}

func (r *CBRepo) Neighbors(ctx context.Context, tsCode string, query CBListQuery) (CBNeighbors, error) {
	tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
	if tsCode == "" {
		return CBNeighbors{}, fmt.Errorf("缺少转债代码")
	}
	joinSQL, where, args := cbListJoinWhere(query)
	orderBy := cbOrderBy(query.SortBy, query.SortDesc)
	sql := `
WITH ranked AS (
  SELECT b.ts_code,
         LAG(b.ts_code) OVER (ORDER BY ` + orderBy + `) AS prev_code,
         LEAD(b.ts_code) OVER (ORDER BY ` + orderBy + `) AS next_code,
         ROW_NUMBER() OVER (ORDER BY ` + orderBy + `) AS rn,
         COUNT(*) OVER () AS total
  FROM cb_basic b
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
		return CBNeighbors{}, err
	}
	out := CBNeighbors{Index: row.RN, Total: row.Total}
	if row.PrevCode != nil {
		out.Prev = *row.PrevCode
	}
	if row.NextCode != nil {
		out.Next = *row.NextCode
	}
	return out, nil
}

func (r *CBRepo) SummarizeHolders(ctx context.Context, tsCode string) ([]CBHolderSummary, error) {
	type row struct {
		EndDate     time.Time
		Top10Ratio  float64
		Top10Amount float64
		HolderCount int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&domain.CBTopHolder{}).
		Select("end_date, COALESCE(SUM(hold_ratio),0) AS top10_ratio, COALESCE(SUM(hold_amount),0) AS top10_amount, COUNT(*) AS holder_count").
		Where("ts_code = ?", tsCode).
		Group("end_date").
		Order("end_date DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]CBHolderSummary, 0, len(rows))
	for _, x := range rows {
		out = append(out, CBHolderSummary{
			EndDate:     x.EndDate,
			Top10Ratio:  x.Top10Ratio,
			Top10Amount: x.Top10Amount,
			HolderCount: x.HolderCount,
		})
	}
	return out, nil
}
