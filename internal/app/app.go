package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/link00000000/gwsn/internal/services"
	"golang.org/x/sync/errgroup"
)

type ServiceContainer struct {
	gmail          services.GmailService
	googleCalendar services.GoogleCalendarService
	notification   services.NotificationService
	systemTray     services.SystemTrayService
}

func (svcs *ServiceContainer) setupServices() error {
	var err error

	errors.Join(err, svcs.gmail.Setup())
	errors.Join(err, svcs.googleCalendar.Setup())
	errors.Join(err, svcs.notification.Setup())
	errors.Join(err, svcs.systemTray.Setup())

	return err
}

func (svcs *ServiceContainer) runServices(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return svcs.gmail.Run(ctx) })
	g.Go(func() error { return svcs.googleCalendar.Run(ctx) })
	g.Go(func() error { return svcs.notification.Run(ctx) })
	g.Go(func() error { return svcs.systemTray.Run(ctx) })

	return g.Wait()
}

func (svcs *ServiceContainer) shutdownServices() error {
	var err error

	errors.Join(err, svcs.gmail.Shutdown())
	errors.Join(err, svcs.googleCalendar.Shutdown())
	errors.Join(err, svcs.notification.Shutdown())
	errors.Join(err, svcs.systemTray.Shutdown())

	return err
}

type Application struct {
	svcs             ServiceContainer
	shutdownRequests chan string
}

var instance *Application = &Application{
	svcs:             ServiceContainer{},
	shutdownRequests: make(chan string),
}

func RegisterGmailService(svc services.GmailService) {
	instance.svcs.gmail = svc
}

func GmailService() services.GmailService {
	return instance.svcs.gmail
}

func RegisterGoogleCalendarService(svc services.GoogleCalendarService) {
	instance.svcs.googleCalendar = svc
}

func GoogleCalendarService() services.GoogleCalendarService {
	return instance.svcs.googleCalendar
}

func RegisterNotificationService(svc services.NotificationService) {
	instance.svcs.notification = svc
}

func NotificationService() services.NotificationService {
	return instance.svcs.notification
}

func RegisterSystemTrayService(svc services.SystemTrayService) {
	instance.svcs.systemTray = svc
}

func SystemTrayService() services.SystemTrayService {
	return instance.svcs.systemTray
}

func Logger() *slog.Logger {
	return slog.Default()
}

func Run(ctx context.Context) error {
	if err := instance.svcs.setupServices(); err != nil {
		return errors.Join(err, instance.svcs.shutdownServices())
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return instance.svcs.runServices(ctx) })
	g.Go(func() error {
		select {
		case reason := <-instance.shutdownRequests:
			Logger().Debug("received shutdown request", "reason", reason)
			cancel()

		case <-ctx.Done():
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return errors.Join(err, instance.svcs.shutdownServices())
	}

	return instance.svcs.shutdownServices()
}

func RequestShutdown(reason string) {
	instance.shutdownRequests <- reason
}
