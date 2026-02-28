package oauth

import (
	"app/config"
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	SCOPES = []string{"profile", "email"}
)

type GoogleAdapter interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
}

type GoogleAdapterImpl struct {
	Conf   *config.Config `inject:"config"`
	OAuth2 *oauth2.Config
}

func (g *GoogleAdapterImpl) Startup() error {
	g.OAuth2 = &oauth2.Config{
		RedirectURL:  g.Conf.OAuth2.RedirectURL,
		ClientID:     g.Conf.OAuth2.GoogleClientID,
		ClientSecret: g.Conf.OAuth2.GoogleClientSecret,
		Scopes:       SCOPES,
		Endpoint:     google.Endpoint,
	}

	return nil
}

func (g *GoogleAdapterImpl) Shutdown() error {
	return nil
}

func (g *GoogleAdapterImpl) AuthCodeURL(state string) string {
	return g.OAuth2.AuthCodeURL(state)
}

func (g *GoogleAdapterImpl) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.OAuth2.Exchange(ctx, code)
}
