package app

import (
	"fmt"
	"io"

	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/ingest"
	"github.com/stock/ashare/internal/repo"
	"github.com/stock/ashare/internal/tushare"
	"gorm.io/gorm"
)

func NewIngestService(cfg config.Config, gdb *gorm.DB, out io.Writer) (*ingest.Service, error) {
	if cfg.TushareToken == "" {
		return nil, fmt.Errorf("未配置 TUSHARE_TOKEN，请在 .env 中填写 Tushare Pro token")
	}
	client := tushare.NewClient(cfg.TushareToken, cfg.TushareURL, cfg.TushareRPM)
	return &ingest.Service{
		Source:      tushare.NewSource(client),
		Instruments: repo.NewInstrumentRepo(gdb),
		Market:      repo.NewMarketRepo(gdb),
		Runs:        repo.NewIngestRunRepo(gdb),
		Out:         out,
	}, nil
}
