package auth

import (
	"context"

	"golang.org/x/oauth2"

	"linuxdo-relay/internal/config"
)

// NewLinuxDoOAuthConfig builds an oauth2.Config for LinuxDo.
func NewLinuxDoOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.LinuxDoClientID,
		ClientSecret: cfg.LinuxDoClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.LinuxDoAuthURL,
			TokenURL: cfg.LinuxDoTokenURL,
		},
		Scopes:      []string{"identify"},
		RedirectURL: cfg.LinuxDoRedirectURL,
	}
}

// ExchangeCode exchanges authorization code for token.
func ExchangeCode(ctx context.Context, conf *oauth2.Config, code string) (*oauth2.Token, error) {
	return conf.Exchange(ctx, code)
}
