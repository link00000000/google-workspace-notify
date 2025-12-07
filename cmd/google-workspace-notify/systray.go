package main

import (
	"context"
	"errors"
	"log"

	"github.com/getlantern/systray"
	"github.com/link00000000/google-workspace-notify/cmd/google-workspace-notify/assets"
)

var ErrSystrayExitRequested = errors.New("exit requested")

func RunSystray(ctx context.Context) error {
	exit := make(chan struct{})

	go systray.Run(func() { onSystrayReady(ctx, exit) }, nil)
	defer systray.Quit()

	select {
	case <-exit:
		return ErrSystrayExitRequested
	case <-ctx.Done():
		close(exit)
	}

	return nil
}

func onSystrayReady(ctx context.Context, exit chan<- struct{}) {
	systray.SetIcon(assets.TrayIcon)
	systray.SetTitle("Google Workspace Notify")

	mSettings := systray.AddMenuItem("Settings", "")
	go runSystrayClickHandlerSettings(ctx, exit, mSettings)

	systray.AddSeparator()

	mExit := systray.AddMenuItem("Exit", "")
	go runSystrayClickHandlerExit(ctx, exit, mExit)
}

func runSystrayClickHandlerSettings(ctx context.Context, onExit chan<- struct{}, m *systray.MenuItem) {
	for {
		select {
		case <-m.ClickedCh:
			log.Println("settings systray menu item clicked")
		case <-ctx.Done():
			return
		}
	}
}

func runSystrayClickHandlerExit(ctx context.Context, exit chan<- struct{}, m *systray.MenuItem) {
	for {
		select {
		case <-m.ClickedCh:
			log.Println("exit systray menu item clicked")
			close(exit)
		case <-ctx.Done():
			return
		}
	}
}
