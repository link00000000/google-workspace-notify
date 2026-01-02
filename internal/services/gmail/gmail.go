package gmail

import (
	"context"
	"time"

	"github.com/link00000000/gwsn/internal/services"
)

type AccountCredentials struct {
	TokenType    string
	AccessToken  string
	RefreshToken string
	Expiry       string
	ExpiresIn    int
}

type Account struct {
	Name  string
	Creds AccountCredentials
}

type gmailService struct {
	pollingInterval time.Duration
	accounts        []Account
}

var _ services.GmailService = (*gmailService)(nil)

func NewService(pollingInterval time.Duration, accounts []Account) *gmailService {
	return &gmailService{
		pollingInterval: pollingInterval,
		accounts:        accounts,
	}
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
