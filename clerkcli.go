package clerkcligo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"

	"golang.org/x/oauth2"
)

type ClerkCli struct {
	Conf      ClerkConf
	browserFn func(url string)
	tokenMgr  ClerkTokenMgr
}

func NewClerkCli(
	conf ClerkConf,
	browserFn func(url string),
	tokenMgr ClerkTokenMgr,
) *ClerkCli {
	c := &ClerkCli{
		Conf:      conf,
		browserFn: browserFn,
		tokenMgr:  tokenMgr,
	}
	return c
}

func (c *ClerkCli) Login(ctx context.Context) error {
	verifier, challenge, err := createPKCEPair()
	if err != nil {
		return err
	}

	state, err := createState()
	if err != nil {
		return err
	}

	authURL, err := c.createAuthorizeURL(challenge, state)
	if err != nil {
		return err
	}

	srv := startServer(state, c.Conf)
	defer srv.shutdown(ctx)
	c.browserFn(authURL)
	code, err := srv.waitForCode(ctx)
	if err != nil {
		return err
	}

	tok, err := exchangeCode(ctx, c.Conf, code, verifier)
	if err != nil {
		return err
	}

	clerkToken := ClerkToken{
		RefreshToken: tok.RefreshToken,
		AccessToken:  tok.AccessToken,
	}

	err = c.tokenMgr.saveTokenFn(clerkToken)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClerkCli) createAuthorizeURL(
	challenge string,
	state string,
) (string, error) {
	authorizeURL := fmt.Sprintf("%s/oauth/authorize", c.Conf.AccountURI)
	authURL, err := url.Parse(authorizeURL)
	if err != nil {
		return "", err
	}

	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("code_challenge_method", "S256")
	q.Set("client_id", c.Conf.ClientID)
	q.Set("redirect_uri", c.Conf.RedirectURI())
	q.Set("scope", "email offline_access profile")
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	authURL.RawQuery = q.Encode()
	return authURL.String(), nil
}

func createPKCEPair() (verifier string, challenge string, err error) {
	b := make([]byte, 64)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func createState() (string, error) {
	state := make([]byte, 24)
	if _, err := rand.Read(state); err != nil {
		return "", err
	}

	stateb64 := base64.RawURLEncoding.EncodeToString(state)
	return stateb64, nil
}

func exchangeCode(
	ctx context.Context,
	conf ClerkConf,
	code string,
	codeVerifier string,
) (*oauth2.Token, error) {
	cfg := oauth2.Config{
		ClientID:    conf.ClientID,
		RedirectURL: conf.RedirectURI(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  conf.AccountURI + "/oauth/authorize",
			TokenURL: conf.AccountURI + "/oauth/token",
		},
	}

	tok, err := cfg.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, err
	}
	return tok, nil
}
