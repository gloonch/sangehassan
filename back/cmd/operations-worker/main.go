package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sangehassan/back/internal/adapters/persistence/postgres"
	"sangehassan/back/internal/config"
	"sangehassan/back/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()
	var provider usecase.SMSProvider = usecase.DisabledSMSProvider{}
	if cfg.SMSProvider == "fake" && cfg.AppEnv != "production" {
		provider = &usecase.FakeSMSProvider{}
	}
	service := usecase.NewOperationsService(db)
	service.ConfigureFinanceAndDocuments(cfg.WorkflowFileDir, provider)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	interval := time.Duration(cfg.WorkerPollSeconds) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		results, runErr := service.RunWorkerOnce(ctx, cfg.NotificationRetrySchedule)
		if runErr != nil {
			slog.Error("worker_cycle_failed", "error", runErr)
		}
		for _, result := range results {
			if result.Error != "" {
				slog.Error("worker_job_failed", "job", result.Job, "affected", result.Affected, "error", result.Error)
			}
		}
		if _, cleanupErr := service.CleanupOrphanDocumentFiles(ctx); cleanupErr != nil {
			slog.Error("orphan_cleanup_failed", "error", cleanupErr)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
