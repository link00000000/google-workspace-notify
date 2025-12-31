package main

import (
	"context"

	"github.com/link00000000/gwsn/internal/app"
	"github.com/link00000000/gwsn/internal/services/gmail"
	"github.com/link00000000/gwsn/internal/services/googlecalendar"
	"github.com/link00000000/gwsn/internal/services/notification"
	"github.com/link00000000/gwsn/internal/services/systemtray"
	"github.com/link00000000/gwsn/internal/services/systemtray/assets"
)

func main() {
	app.RegisterGmailService(gmail.NewService())
	app.RegisterGoogleCalendarService(googlecalendar.NewService())
	app.RegisterNotificationService(notification.NewBeeepNotificationService("Google Workspace Notify"))
	app.RegisterSystemTrayService(systemtray.NewSystraySystemTrayService("Google Workspace Notify", assets.TrayIcon))

	if err := app.Run(context.Background()); err != nil {
		panic(err)
	}
}
