package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stock/ashare/internal/app"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/internal/httpapi"
	"github.com/stock/ashare/internal/ingest"
)

func cmdWeb(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gdb, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	var runner *ingest.Runner
	if svc, err := app.NewIngestService(cfg, gdb, os.Stdout); err != nil {
		fmt.Printf("采集 API 未启用: %v\n", err)
	} else {
		runner = ingest.NewRunner(svc)
		fmt.Printf("Tushare 限流 %d 次/分钟\n", cfg.TushareRPM)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewMux(gdb, cfg, runner),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("API  http://%s\n", cfg.HTTPAddr)
		fmt.Println("前端请另开终端:  cd web && npm run dev")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
