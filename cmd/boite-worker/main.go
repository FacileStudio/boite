package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FacileStudio/boite/cmd/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("boite-worker", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := worker.LoadConfig(os.Getenv)
	if err != nil {
		return err
	}
	ln, err := worker.Listen(cfg.Sock, cfg.SockMode, cfg.SockGID)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	slog.Info("boite-worker listening", "sock", cfg.Sock, "mode", cfg.SockMode.String(), "gid", cfg.SockGID, "host", cfg.Host)
	handler := worker.NewServer(worker.NewBackend(cfg.Host), cfg.Token).Handler()
	return worker.Serve(ctx, ln, handler)
}
