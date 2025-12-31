package notification

import (
	"context"

	"github.com/gen2brain/beeep"
	"github.com/link00000000/gwsn/internal/services"
)

type beeepNotificationService struct {
	appName string
}

var _ services.NotificationService = (*beeepNotificationService)(nil)

func NewBeeepNotificationService(appName string) *beeepNotificationService {
	return &beeepNotificationService{
		appName: appName,
	}
}

func (svc *beeepNotificationService) Setup() error {
	beeep.AppName = svc.appName

	return nil
}

func (*beeepNotificationService) Run(ctx context.Context) error {
	return nil
}

func (*beeepNotificationService) Shutdown() error {
	return nil
}

func (*beeepNotificationService) Notify(title, body string) {
	beeep.Notify(title, body, "")
}

func (*beeepNotificationService) NotifyWithIcon(title, body string, icon []byte) {
	beeep.Notify(title, body, icon)
}
