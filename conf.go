package clerkcligo

import (
	"fmt"
)

type ClerkConf struct {
	AccountURI   string
	RedirectPort int
	RedirectIP   string
	ClientID     string
}

func (c *ClerkConf) RedirectURI() string {
	return fmt.Sprintf("http://%s:%d/callback", c.RedirectIP, c.RedirectPort)
}

func (c *ClerkConf) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.RedirectIP, c.RedirectPort)
}
