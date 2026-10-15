package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stock/ashare/internal/app"
	"github.com/stock/ashare/internal/applog"
	"github.com/stock/ashare/internal/config"
	"github.com/stock/ashare/internal/db"
	"github.com/stock/ashare/internal/httpapi"
	"github.com/stock/ashare/internal/ingest"
)

func cmdWeb(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gdb, err := db.ConnectWithLog(ctx, cfg.DatabaseURL, applog.GormTraceLevel(cfg.Log.Level))
	if err != nil {
		return err
	}
	defer db.Close(gdb)

	var runner *ingest.Runner
	if svc, err := app.NewIngestService(cfg, gdb); err != nil {
		slog.Warn("ingest API disabled", "component", "web", "err", err)
	} else {
		runner = ingest.NewRunner(svc)
		slog.Info("tushare rate limit",
			"component", "web",
			"rpm", cfg.TushareRPM,
			"apis", cfg.TushareAPIs,
			"gap", cfg.MinInterval().Round(time.Millisecond).String(),
		)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewMux(gdb, cfg, runner),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("api listening",
			"component", "web",
			"addr", cfg.HTTPAddr,
			"framework", "gin",
		)
		fmt.Println("前端请另开终端:  cd web && pnpm dev")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("api shutting down", "component", "web")
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
