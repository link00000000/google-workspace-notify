package config

import (
	"slices"
	"time"
)

type InMemoryConfigProvider struct {
	cfg *InMemoryConfig
}

var _ ConfigProvider = (*InMemoryConfigProvider)(nil)

type GmailAccountInMemoryConfig struct {
	Name         *string
	TokenType    *string
	AccessToken  *string
	RefreshToken *string
	Expiry       *string
	ExpiresIn    *int
}

type GmailInMemoryConfig struct {
	Accounts        *[]GmailAccountInMemoryConfig
	PollingInterval *time.Duration
}

type InMemoryConfig struct {
	Gmail *GmailInMemoryConfig
}

func NewInMemoryConfigProvider(cfg *InMemoryConfig) *InMemoryConfigProvider {
	return &InMemoryConfigProvider{
		cfg: cfg,
	}
}

func (p *InMemoryConfigProvider) Apply(cfg *Config) error {
	if p.cfg.Gmail != nil {
		if p.cfg.Gmail.Accounts != nil {
			for _, acc := range *p.cfg.Gmail.Accounts {
				var targetAccount *GmailAccountConfig

				idx := slices.IndexFunc(cfg.Gmail.Accounts, func(a GmailAccountConfig) bool { return a.Name == *acc.Name })
				if idx == -1 {
					cfg.Gmail.Accounts = append(cfg.Gmail.Accounts, GmailAccountConfig{})
					targetAccount = &cfg.Gmail.Accounts[len(cfg.Gmail.Accounts)-1]
				} else {
					targetAccount = &cfg.Gmail.Accounts[idx]
				}

				applyProp(&targetAccount.Name, acc.Name)
				applyProp(&targetAccount.TokenType, acc.TokenType)
				applyProp(&targetAccount.AccessToken, acc.AccessToken)
				applyProp(&targetAccount.RefreshToken, acc.RefreshToken)
				applyProp(&targetAccount.Expiry, acc.Expiry)
				applyProp(&targetAccount.ExpiresIn, acc.ExpiresIn)
			}
		}

		applyProp(&cfg.Gmail.PollingInterval, p.cfg.Gmail.PollingInterval)
	}

	return nil
}
