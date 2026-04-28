package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backup-system/config"
	"backup-system/internal/monitor"
	"backup-system/internal/utils"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := utils.NewLogger()
	service := monitor.NewService(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil {
		logger.Fatalf("monitor failed: %v", err)
	}
}
