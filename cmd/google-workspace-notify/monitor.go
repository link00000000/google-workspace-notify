package main

import (
	"context"
	"log/slog"

	"github.com/link00000000/google-workspace-notify/src/monitor"
)

func RunMonitor(ctx context.Context) error {
	m := monitor.NewMonitor()

	go m.Run()
	defer m.Stop()

	for {
		select {
		case <-m.CalendarReminder():
			slog.Info("recieved calendar reminder") // TODO: add attrs
			// TODO: notify new calendar reminder
		case <-m.Email():
			slog.Info("recieved email") // TODO: add attrs
			// TODO: notify new email
		case <-ctx.Done():
			return nil
		}
	}
}
