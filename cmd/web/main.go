package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"backup-system/config"
	"backup-system/internal/web"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv, err := web.NewServer(cfg)
	if err != nil {
		log.Fatalf("web server init: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		log.Println("shutting down web server")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("web server: %v", err)
		}
	}
}
