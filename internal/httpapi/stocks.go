package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock/ashare/internal/dates"
	"github.com/stock/ashare/internal/repo"
)

type stockListItemDTO struct {
	TSCode      string   `json:"ts_code"`
	Symbol      string   `json:"symbol"`
	Name        string   `json:"name"`
	Exchange    string   `json:"exchange"`
	IsST        bool     `json:"is_st"`
	Status      string   `json:"status"`
	ListDate    string   `json:"list_date,omitempty"`
	DelistDate  string   `json:"delist_date,omitempty"`
	QuoteDate   string   `json:"quote_date,omitempty"`
	PreClose    *float64 `json:"pre_close,omitempty"`
	Close       *float64 `json:"close,omitempty"`
	Change      *float64 `json:"change,omitempty"`
	PctChg      *float64 `json:"pct_chg,omitempty"`
	Volume      *int64   `json:"volume,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	IsLimitUp   *bool    `json:"is_limit_up,omitempty"`
	IsLimitDown *bool    `json:"is_limit_down,omitempty"`
	IsPaused    *bool    `json:"is_paused,omitempty"`
}

func parseStockListQuery(c *gin.Context) (repo.StockListQuery, error) {
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
			return repo.StockListQuery{}, err
		}
		tradeDate = t
	}
	return repo.StockListQuery{
		Q:          c.Query("q"),
		Limit:      limit,
		Offset:     offset,
		ListedOnly: listedOnly,
		TradeDate:  tradeDate,
		SortBy:     sortBy,
		SortDesc:   sortDesc,
	}, nil
}

func (h *Handler) stockList(c *gin.Context) {
	query, err := parseStockListQuery(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	items, total, err := h.instruments.SearchWithQuote(c.Request.Context(), query)
	if err != nil {
		abortJSON(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]stockListItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toStockListRowDTO(it))
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": out,
	})
}

func (h *Handler) stockNeighbors(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if code == "" {
		abortJSON(c, http.StatusBadRequest, "缺少股票代码")
		return
	}
	query, err := parseStockListQuery(c)
	if err != nil {
		abortJSON(c, http.StatusBadRequest, err.Error())
		return
	}
	query.Limit = 0
	query.Offset = 0
	n, err := h.instruments.Neighbors(c.Request.Context(), code, query)
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

func toStockListRowDTO(it repo.StockListRow) stockListItemDTO {
	dto := stockListItemDTO{
		TSCode:      it.TSCode,
		Symbol:      it.Symbol,
		Name:        it.Name,
		Exchange:    it.Exchange,
		IsST:        it.IsST,
		Status:      it.ListStatus,
		ListDate:    dateISO(it.ListDate),
		DelistDate:  dateISO(it.DelistDate),
		PreClose:    it.PreClose,
		Close:       it.Close,
		Change:      it.Change,
		PctChg:      it.PctChg,
		Volume:      it.Volume,
		Amount:      it.Amount,
		IsLimitUp:   it.IsLimitUp,
		IsLimitDown: it.IsLimitDown,
		IsPaused:    it.IsPaused,
	}
	if it.QuoteDate != nil {
		dto.QuoteDate = dates.ISO(*it.QuoteDate)
	}
	return dto
}
