package clerkcligo

import (
	"fmt"

	"golang.org/x/oauth2"
)

type ClerkConf struct {
	AccountURI   string
	RedirectPort int
	RedirectIP   string
	ClientID     string
	Scopes       []string
}

func (c *ClerkConf) RedirectURL() string {
	return fmt.Sprintf("http://%s:%d/callback", c.RedirectIP, c.RedirectPort)
}

func (c *ClerkConf) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.RedirectIP, c.RedirectPort)
}

func (c *ClerkConf) toOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: c.RedirectURL(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.AccountURI + "/oauth/authorize",
			TokenURL: c.AccountURI + "/oauth/token",
		},
		Scopes: c.Scopes,
	}
}
