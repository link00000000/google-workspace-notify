package gmail

import (
	"context"

	"github.com/link00000000/gwsn/internal/services"
)

type gmailService struct{}

var _ services.GmailService = (*gmailService)(nil)

func NewService() *gmailService {
	return &gmailService{}
}

func (*gmailService) Setup() error {
	return nil
}

func (*gmailService) Run(ctx context.Context) error {
	return nil
}

func (*gmailService) Shutdown() error {
	return nil
}
