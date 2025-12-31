package googlecalendar

import (
	"context"

	"github.com/link00000000/gwsn/internal/services"
)

type googleCalendarService struct{}

var _ services.GoogleCalendarService = (*googleCalendarService)(nil)

func NewService() *googleCalendarService {
	return &googleCalendarService{}
}

func (*googleCalendarService) Setup() error {
	return nil
}

func (*googleCalendarService) Run(ctx context.Context) error {
	return nil
}

func (*googleCalendarService) Shutdown() error {
	return nil
}
