package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/domain"
	"github.com/stock/ashare/internal/repo"
	"gorm.io/gorm"
)

type cbListItemDTO struct {
	TSCode            string   `json:"ts_code"`
	BondShortName     string   `json:"bond_short_name"`
	BondFullName      string   `json:"bond_full_name"`
	CBType            string   `json:"cb_type"`
	StkCode           string   `json:"stk_code"`
	StkShortName      string   `json:"stk_short_name"`
	Exchange          string   `json:"exchange"`
	ListDate          string   `json:"list_date,omitempty"`
	DelistDate        string   `json:"delist_date,omitempty"`
	ConvPrice         *float64 `json:"conv_price,omitempty"`
	CouponRate        *float64 `json:"coupon_rate,omitempty"`
	IssueRating       string   `json:"issue_rating,omitempty"`
	NewestRating      string   `json:"newest_rating,omitempty"`
	IssueSize         *float64 `json:"issue_size,omitempty"`  // 初始规模（元）
	RemainSize        *float64 `json:"remain_size,omitempty"` // 剩余规模（元）
	FloatSize         *float64 `json:"float_size,omitempty"`  // 流通规模（元）
	Top10Ratio        *float64 `json:"top10_ratio,omitempty"` // 最近一期十大合计占比 %
	QuoteDate         string   `json:"quote_date,omitempty"`
	PreClose          *float64 `json:"pre_close,omitempty"`
	Close             *float64 `json:"close,omitempty"`
	Change            *float64 `json:"change,omitempty"`
	PctChg            *float64 `json:"pct_chg,omitempty"`
	Volume            *float64 `json:"volume,omitempty"`
	Amount            *float64 `json:"amount,omitempty"`
	CBOverRate        *float64 `json:"cb_over_rate,omitempty"`        // 转股溢价率 %
	MaturityCallPrice *float64 `json:"maturity_call_price,omitempty"` // 到期赎回价
}

type cbDetailDTO struct {
	TSCode            string   `json:"ts_code"`
	BondFullName      string   `json:"bond_full_name"`
	BondShortName     string   `json:"bond_short_name"`
	CBCode            string   `json:"cb_code"`
	CBType            string   `json:"cb_type"`
	StkCode           string   `json:"stk_code"`
	StkShortName      string   `json:"stk_short_name"`
	Maturity          *float64 `json:"maturity,omitempty"`
	Par               *float64 `json:"par,omitempty"`
	IssuePrice        *float64 `json:"issue_price,omitempty"`
	IssueSize         *float64 `json:"issue_size,omitempty"`
	RemainSize        *float64 `json:"remain_size,omitempty"`
	ValueDate         string   `json:"value_date,omitempty"`
	MaturityDate      string   `json:"maturity_date,omitempty"`
	RateType          string   `json:"rate_type,omitempty"`
	CouponRate        *float64 `json:"coupon_rate,omitempty"`
	AddRate           *float64 `json:"add_rate,omitempty"`
	PayPerYear        *int     `json:"pay_per_year,omitempty"`
	ListDate          string   `json:"list_date,omitempty"`
	DelistDate        string   `json:"delist_date,omitempty"`
	Exchange          string   `json:"exchange"`
	ConvStartDate     string   `json:"conv_start_date,omitempty"`
	ConvEndDate       string   `json:"conv_end_date,omitempty"`
	ConvStopDate      string   `json:"conv_stop_date,omitempty"`
	FirstConvPrice    *float64 `json:"first_conv_price,omitempty"`
	ConvPrice         *float64 `json:"conv_price,omitempty"`
	RateClause        string   `json:"rate_clause,omitempty"`
	PutClause         string   `json:"put_clause,omitempty"`
	MaturityCallPrice *float64 `json:"maturity_call_price,omitempty"`
	CallClause        string   `json:"call_clause,omitempty"`
	ResetClause       string   `json:"reset_clause,omitempty"`
	ConvClause        string   `json:"conv_clause,omitempty"`
	Guarantor         string   `json:"guarantor,omitempty"`
	GuaranteeType     string   `json:"guarantee_type,omitempty"`
	IssueRating       string   `json:"issue_rating,omitempty"`
	NewestRating      string   `json:"newest_rating,omitempty"`
	RatingComp        string   `json:"rating_comp,omitempty"`
}

type cbHolderDTO struct {
	EndDate    string   `json:"end_date"`
	HolderRank int      `json:"holder_rank"`
	HolderName string   `json:"holder_name"`
	HoldAmount *float64 `json:"hold_amount"`
	HoldRatio  *float64 `json:"hold_ratio"`
}

type cbHolderSummaryDTO struct {
	EndDate     string  `json:"end_date"`
	Top10Ratio  float64 `json:"top10_ratio"`
	Top10Amount float64 `json:"top10_amount"`
	HolderCount int64   `json:"holder_count"`
}

func parseCBListQuery(c *gin.Context) (repo.CBListQuery, error) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	listedOnly := c.Query("listed_only") == "1" || c.Query("listed_only") == "true"
	sortBy := strings.TrimSpace(c.Query("sort_by"))
	sortDir := strings.ToLower(strings.TrimSpace(c.Query("sort_dir")))
	sortDesc := sortDir == "" || sortDir == "desc" || sortDir == "descending"
	if sortDir == "asc" || sortDir == "ascending" {
		sortDesc = false
	}
	var tradeDate time.Time
	if s := strings.TrimSpace(c.Query("trade_date")); s != "" {
		t, err := dates.Parse(s)
		if err != nil {
			return repo.CBListQuery{}, err
		}
		tradeDate = t
	}
	return repo.CBListQuery{
		Q:          c.Query("q"),
		Limit:      limit,
		Offset:     offset,
		ListedOnly: listedOnly,
		TradeDate:  tradeDate,
		SortBy:     sortBy,
		SortDesc:   sortDesc,
	}, nil
}

func (h *Handler) cbList(c *gin.Context) {
	if h.cb == nil {
		abortJSON(c, http.StatusServiceUnavailable, "可转债仓储未启用")
		return
	}
	query, err := parseCBListQuery(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	items, total, err := h.cb.SearchWithQuote(c.Request.Context(), query)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]cbListItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toCBListRowDTO(it))
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": out,
	})
}

func (h *Handler) cbNeighbors(c *gin.Context) {
	if h.cb == nil {
		abortJSON(c, http.StatusServiceUnavailable, "可转债仓储未启用")
		return
	}
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少转债代码")
		return
	}
	query, err := parseCBListQuery(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	// 邻居接口不需要分页参数
	query.Limit = 0
	query.Offset = 0
	n, err := h.cb.Neighbors(c.Request.Context(), code, query)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"prev":  n.Prev,
		"next":  n.Next,
		"index": n.Index,
		"total": n.Total,
	})
}

func (h *Handler) cbDetail(c *gin.Context) {
	if h.cb == nil {
		abortJSON(c, http.StatusServiceUnavailable, "可转债仓储未启用")
		return
	}
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少转债代码")
		return
	}
	basic, err := h.cb.GetByTSCode(c.Request.Context(), code)
	if err == gorm.ErrRecordNotFound {
		abortJSON(c, http.StatusNotFound, "未找到可转债 "+code+"，请先 ingest cb-basic")
		return
	}
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	periods, err := h.cb.ListHolderPeriods(c.Request.Context(), code)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	summaries, err := h.cb.SummarizeHolders(c.Request.Context(), code)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	periodStr := strings.TrimSpace(c.Query("period"))
	var period time.Time
	if periodStr != "" {
		period, err = dates.Parse(periodStr)
		if err != nil {
			abortJSON(c, http.StatusBadRequest, err.Error())
			return
		}
	} else if len(periods) > 0 {
		period = periods[0]
	}

	var holders []domain.CBTopHolder
	if !period.IsZero() {
		holders, err = h.cb.ListTopHolders(c.Request.Context(), code, period)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	periodISOs := make([]string, 0, len(periods))
	for _, p := range periods {
		periodISOs = append(periodISOs, dates.ISO(p))
	}
	holderDTOs := make([]cbHolderDTO, 0, len(holders))
	for _, row := range holders {
		holderDTOs = append(holderDTOs, toCBHolderDTO(row))
	}
	sumDTOs := make([]cbHolderSummaryDTO, 0, len(summaries))
	for _, s := range summaries {
		sumDTOs = append(sumDTOs, toCBHolderSummaryDTO(s))
	}

	selected := ""
	if !period.IsZero() {
		selected = dates.ISO(period)
	}
	c.JSON(http.StatusOK, gin.H{
		"basic":            toCBDetailDTO(basic),
		"holder_periods":   periodISOs,
		"selected_period":  selected,
		"holders":          holderDTOs,
		"holder_summaries": sumDTOs,
	})
}

type cbReplaceHolderItem struct {
	HolderRank int     `json:"holder_rank"`
	HolderName string  `json:"holder_name"`
	HoldAmount float64 `json:"hold_amount"`
	HoldRatio  float64 `json:"hold_ratio"`
}

type cbReplaceHoldersReq struct {
	Period  string                `json:"period"`
	Holders []cbReplaceHolderItem `json:"holders"`
}

func (h *Handler) cbReplaceHolders(c *gin.Context) {
	if h.cb == nil {
		abortJSON(c, http.StatusServiceUnavailable, "可转债仓储未启用")
		return
	}
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少转债代码")
		return
	}
	var req cbReplaceHoldersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		abortJSON(c, http.StatusBadRequest, "请求体无效")
		return
	}
	period, err := dates.Parse(strings.TrimSpace(req.Period))
	if err != nil {
		abortJSON(c, http.StatusBadRequest, "period 无效，期望 YYYY-MM-DD")
		return
	}
	if len(req.Holders) == 0 {
		abortJSON(c, http.StatusBadRequest, "holders 不能为空")
		return
	}
	items := make([]domain.CBTopHolder, 0, len(req.Holders))
	for i, row := range req.Holders {
		name := strings.TrimSpace(row.HolderName)
		if name == "" {
			abortJSON(c, http.StatusBadRequest, "holder_name 不能为空")
			return
		}
		rank := row.HolderRank
		if rank <= 0 {
			rank = i + 1
		}
		amt, rat := row.HoldAmount, row.HoldRatio
		items = append(items, domain.CBTopHolder{
			TSCode:     code,
			EndDate:    period,
			HolderRank: rank,
			HolderName: name,
			HoldAmount: &amt,
			HoldRatio:  &rat,
		})
	}
	if _, err := h.cb.GetByTSCode(c.Request.Context(), code); err == gorm.ErrRecordNotFound {
		abortJSON(c, http.StatusNotFound, "未找到可转债 "+code+"，请先 ingest cb-basic")
		return
	} else if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.cb.ReplaceTopHolders(c.Request.Context(), code, period, items); err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	holders, err := h.cb.ListTopHolders(c.Request.Context(), code, period)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]cbHolderDTO, 0, len(holders))
	for _, row := range holders {
		out = append(out, toCBHolderDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{
		"period":  dates.ISO(period),
		"count":   len(out),
		"holders": out,
	})
}

func dateISO(t *time.Time) string {
	if t == nil {
		return ""
	}
	return dates.ISO(*t)
}

func toCBListRowDTO(it repo.CBListRow) cbListItemDTO {
	dto := cbListItemDTO{
		TSCode:            it.TSCode,
		BondShortName:     it.BondShortName,
		BondFullName:      it.BondFullName,
		CBType:            it.CBType,
		StkCode:           it.StkCode,
		StkShortName:      it.StkShortName,
		Exchange:          it.Exchange,
		ListDate:          dateISO(it.ListDate),
		DelistDate:        dateISO(it.DelistDate),
		ConvPrice:         it.ConvPrice,
		CouponRate:        it.CouponRate,
		IssueRating:       it.IssueRating,
		NewestRating:      it.NewestRating,
		IssueSize:         it.IssueSize,
		RemainSize:        it.RemainSize,
		FloatSize:         it.FloatSize,
		Top10Ratio:        it.Top10Ratio,
		PreClose:          it.PreClose,
		Close:             it.Close,
		Change:            it.Change,
		PctChg:            it.PctChg,
		Volume:            it.Volume,
		Amount:            it.Amount,
		CBOverRate:        it.CBOverRate,
		MaturityCallPrice: it.MaturityCallPrice,
	}
	if it.QuoteDate != nil {
		dto.QuoteDate = dates.ISO(*it.QuoteDate)
	}
	return dto
}

func toCBDetailDTO(it domain.CBBasic) cbDetailDTO {
	return cbDetailDTO{
		TSCode:            it.TSCode,
		BondFullName:      it.BondFullName,
		BondShortName:     it.BondShortName,
		CBCode:            it.CBCode,
		CBType:            it.CBType,
		StkCode:           it.StkCode,
		StkShortName:      it.StkShortName,
		Maturity:          it.Maturity,
		Par:               it.Par,
		IssuePrice:        it.IssuePrice,
		IssueSize:         it.IssueSize,
		RemainSize:        it.RemainSize,
		ValueDate:         dateISO(it.ValueDate),
		MaturityDate:      dateISO(it.MaturityDate),
		RateType:          it.RateType,
		CouponRate:        it.CouponRate,
		AddRate:           it.AddRate,
		PayPerYear:        it.PayPerYear,
		ListDate:          dateISO(it.ListDate),
		DelistDate:        dateISO(it.DelistDate),
		Exchange:          it.Exchange,
		ConvStartDate:     dateISO(it.ConvStartDate),
		ConvEndDate:       dateISO(it.ConvEndDate),
		ConvStopDate:      dateISO(it.ConvStopDate),
		FirstConvPrice:    it.FirstConvPrice,
		ConvPrice:         it.ConvPrice,
		RateClause:        it.RateClause,
		PutClause:         it.PutClause,
		MaturityCallPrice: it.MaturityCallPrice,
		CallClause:        it.CallClause,
		ResetClause:       it.ResetClause,
		ConvClause:        it.ConvClause,
		Guarantor:         it.Guarantor,
		GuaranteeType:     it.GuaranteeType,
		IssueRating:       it.IssueRating,
		NewestRating:      it.NewestRating,
		RatingComp:        it.RatingComp,
	}
}

func toCBHolderDTO(row domain.CBTopHolder) cbHolderDTO {
	return cbHolderDTO{
		EndDate:    dates.ISO(row.EndDate),
		HolderRank: row.HolderRank,
		HolderName: row.HolderName,
		HoldAmount: row.HoldAmount,
		HoldRatio:  row.HoldRatio,
	}
}

func toCBHolderSummaryDTO(s repo.CBHolderSummary) cbHolderSummaryDTO {
	return cbHolderSummaryDTO{
		EndDate:     dates.ISO(s.EndDate),
		Top10Ratio:  s.Top10Ratio,
		Top10Amount: s.Top10Amount,
		HolderCount: s.HolderCount,
	}
}
