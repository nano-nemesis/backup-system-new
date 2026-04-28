package main

import (
	"context"
	"log"

	"backup-system/config"
	"backup-system/internal/backup"
	"backup-system/internal/utils"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := utils.NewLogger()
	runner := backup.NewRunner(cfg, logger)

	if err := runner.Run(context.Background()); err != nil {
		logger.Fatalf("backup run failed: %v", err)
	}

	logger.Println("backup run completed successfully")
}
