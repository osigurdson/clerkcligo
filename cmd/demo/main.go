package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/osigurdson/workoscligo"
)

func main() {
	ctx := context.Background()
	/*
		original := workoscligo.WorkOSConf{
			AuthKitURI:   "https://your-org.authkit.app",
			RedirectIP:   "127.0.0.1",
			RedirectPort: 21222,
			ClientID:     "client_123",
			Scopes:       []string{"openid", "profile", "email", "offline_access"},
		}
	*/

	conf := workoscligo.WorkOSConf{
		AuthKitURI:   "https://surprising-cliff-63-staging.authkit.app",
		RedirectIP:   "127.0.0.1",
		RedirectPort: 21222,
		ClientID:     "client_01KHQ0D7B6Y3J3S7AXH9RQYX85",
		Scopes:       []string{"openid", "profile", "email", "offline_access"},
	}

	var token *workoscligo.WorkOSToken

	mgr, err := workoscligo.NewWorkOSTokenMgr(
		func(newToken workoscligo.WorkOSToken) error {
			token = &newToken
			fmt.Printf("token: %+v\n", token)
			return nil
		},
		func() (workoscligo.WorkOSToken, error) {
			if token == nil {
				return workoscligo.WorkOSToken{}, fmt.Errorf("WorkOS token not found")
			}
			return *token, nil
		},
	)

	browser := func(url string) {
		exec.Command("xdg-open", url).Start()
	}

	workosCli := workoscligo.NewWorkOSCli(
		conf,
		browser,
		mgr,
	)

	err = workosCli.Login(ctx)
	if err != nil {
		panic(err)
	}

	client := workosCli.NewHttpClient(ctx)
	res, err := client.Get("http://localhost:5000/api/v1/stores")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
}
