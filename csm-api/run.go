package main

import (
	"context"
	"csm-api/config"
	"csm-api/utils"
	"fmt"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"time"
)

func runWeb(ctx context.Context, cfg *config.Config, safeDb, timesheetDb *sqlx.DB) error {
	// 포트/도메인 리스너
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Domain, cfg.Port))
	if err != nil {
		return utils.CustomMessageErrorf("net.Listen", err)
	}
	url := fmt.Sprintf("http://%s", l.Addr().String())
	log.Printf("Listening at %s\n", url)

	// mux/route
	mux, err := newMux(ctx, safeDb, timesheetDb)
	if err != nil {
		return utils.CustomMessageErrorf("newMux", err)
	}

	server := NewServer(l, mux)

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return server.Run(ctx) })

	// 종료 신호 대기 및 graceful shutdown
	select {
	case <-ctx.Done():
		fmt.Println("\nShutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return utils.CustomMessageErrorf("server graceful shutdown", err)
		}
		log.Println("Server exited normally.")
	}

	return eg.Wait()
}

func runSchedule(ctx context.Context, safeDb, timesheetDb *sqlx.DB) error {
	apiCfg, err := config.GetApiConfig()
	if err != nil {
		return utils.CustomMessageErrorf("config.ApiConfig", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		return utils.CustomMessageErrorf("config.Config", err)
	}

	scheduler, err := NewScheduler(safeDb, apiCfg, cfg, timesheetDb)
	if err != nil {
		return utils.CustomMessageErrorf("NewScheduler", err)
	}
	// 스케줄러는 Run만 실행, 종료 신호는 내부에서 ctx.Done()으로 처리
	return scheduler.Run(ctx)
}
