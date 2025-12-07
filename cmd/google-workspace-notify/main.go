package main

import (
	"context"
	"log/slog"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("starting RunSystray")
		if err := RunSystray(ctx); err != nil {
			if err == ErrSystrayExitRequested {
				cancel()
			} else {
				slog.Error("RunSystray completed with unhandled error", "error", err)
				return err
			}
		}

		slog.Info("RunSystray completed without error")
		return nil
	})

	g.Go(func() error {
		slog.Info("starting RunHttpServer")
		if err := RunHttpServer(ctx); err != nil {
			slog.Error("RunHttpServer completed with unhandled error", "error", err)
			return err
		}

		slog.Info("RunHttpServer completed without error")
		return nil
	})

	g.Go(func() error {
		slog.Info("starting RunMonitor")
		if err := RunMonitor(ctx); err != nil {
			slog.Error("RunMonitor completed with unhandled error", "error", err)
			return err
		}

		slog.Info("RunMonitor completed without error")
		return nil
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
