package systemtray

import (
	"context"

	"github.com/getlantern/systray"
	"github.com/link00000000/gwsn/internal/app"
	"github.com/link00000000/gwsn/internal/services"
)

type systraySystemTrayService struct {
	title    string
	trayIcon []byte
}

var _ services.SystemTrayService = (*systraySystemTrayService)(nil)

func NewSystraySystemTrayService(title string, trayIcon []byte) *systraySystemTrayService {
	return &systraySystemTrayService{
		title:    title,
		trayIcon: trayIcon,
	}
}

func (*systraySystemTrayService) Setup() error {
	return nil
}

func (svc *systraySystemTrayService) Run(ctx context.Context) error {
	systray.Run(func() {
		systray.SetIcon(svc.trayIcon)
		systray.SetTitle(svc.title)

		settingsEntry := systray.AddMenuItem("Settings", "")
		systray.AddSeparator()
		exitEntry := systray.AddMenuItem("Exit", "")

		for {
			select {
			case <-settingsEntry.ClickedCh:
				panic("not implemented")

			case <-exitEntry.ClickedCh:
				app.RequestShutdown("exit requested by user")

			case <-ctx.Done():
				systray.Quit()
				return
			}
		}
	}, nil)

	return nil
}

func (*systraySystemTrayService) Shutdown() error {
	return nil
}
