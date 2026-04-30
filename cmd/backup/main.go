package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backup-system/config"
	"backup-system/internal/backup"
	"backup-system/internal/utils"
)

func main() {
	once := flag.Bool("once", false, "run a single backup immediately and exit")
	flag.Parse()

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := utils.NewLogger()
	runner := backup.NewRunner(cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// One-shot mode: run immediately and exit.
	if *once {
		if err := runner.Run(ctx); err != nil {
			logger.Fatalf("backup run failed: %v", err)
		}
		logger.Println("backup run completed")
		return
	}

	// Daemon mode: run on a fixed schedule aligned to midnight in the configured timezone.
	loc, err := time.LoadLocation(cfg.BackupTimezone)
	if err != nil {
		logger.Fatalf("invalid BACKUP_TIMEZONE %q: %v", cfg.BackupTimezone, err)
	}

	interval := cfg.BackupIntervalHours
	if interval <= 0 {
		interval = 3
	}

	logger.Printf("backup scheduler started: every %dh, timezone %s", interval, cfg.BackupTimezone)
	logger.Printf("daily schedule: %s", scheduleDescription(interval))

	for {
		next := nextScheduledTime(time.Now(), loc, interval)
		wait := time.Until(next)
		logger.Printf("next backup at %s (in %s)",
			next.In(loc).Format("2006-01-02 15:04:05 MST"),
			wait.Round(time.Second),
		)

		select {
		case <-ctx.Done():
			logger.Println("backup scheduler stopped")
			return
		case <-time.After(wait):
		}

		logger.Println("starting scheduled backup")
		if err := runner.Run(ctx); err != nil {
			logger.Printf("backup run error: %v", err)
		}
	}
}

// nextScheduledTime returns the next slot that is strictly after now, aligned
// to midnight in loc with the given interval (hours). For interval=3 the slots
// are 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00.
func nextScheduledTime(now time.Time, loc *time.Location, intervalHours int) time.Time {
	local := now.In(loc)
	y, m, d := local.Date()

	for h := 0; h < 24; h += intervalHours {
		t := time.Date(y, m, d, h, 0, 0, 0, loc)
		if t.After(now) {
			return t
		}
	}

	// All today's slots are past — first slot tomorrow.
	return time.Date(y, m, d+1, 0, 0, 0, 0, loc)
}

func scheduleDescription(intervalHours int) string {
	var slots []string
	for h := 0; h < 24; h += intervalHours {
		slots = append(slots, time.Date(0, 1, 1, h, 0, 0, 0, time.UTC).Format("15:04"))
	}
	result := ""
	for i, s := range slots {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
