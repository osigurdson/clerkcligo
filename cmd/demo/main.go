package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/osigurdson/clerkcligo"
)

func main() {
	ctx := context.Background()
	conf := clerkcligo.ClerkConf{
		AccountURI:   "https://top-haddock-51.clerk.accounts.dev",
		RedirectIP:   "127.0.0.1",
		RedirectPort: 21222,
		ClientID:     "uqoyQTDEq3yLqJeH",
		Scopes:       []string{"email", "profile", "offline_access"},
	}

	var token *clerkcligo.ClerkToken

	mgr, err := clerkcligo.NewClerkTokenMgr(
		func(newToken clerkcligo.ClerkToken) error {
			token = &newToken
			return nil
		},
		func() (clerkcligo.ClerkToken, error) {
			if token == nil {
				return clerkcligo.ClerkToken{}, fmt.Errorf("Clerk token not found")
			}
			return *token, nil
		},
	)

	browser := func(url string) {
		exec.Command("xdg-open", url).Start()
	}

	clerkCli := clerkcligo.NewClerkCli(
		conf,
		browser,
		mgr,
	)

	err = clerkCli.Login(ctx)
	if err != nil {
		panic(err)
	}

	client := clerkCli.NewHttpClient(ctx)
	res, err := client.Get("http://localhost.com/api/v1/stores")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
}
