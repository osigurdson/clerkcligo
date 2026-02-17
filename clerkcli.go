package clerkcligo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type ClerkCli struct {
	Conf      ClerkConf
	browserFn func(url string)
	tokenMgr  ClerkTokenMgr
	debug     bool
	mu        sync.Mutex
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
		debug:     true,
	}
	return c
}

func (c *ClerkCli) NewHttpClient(ctx context.Context) *http.Client {
	transport := &AuthTransport{
		Auth: c,
		Base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
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
	c.printdbg(fmt.Sprintf("Login: Step 1 - Create /authorize url. URL: %s\n", authURL))

	srv := startServer(state, c.Conf)
	defer srv.shutdown(ctx)

	c.printdbg("Login: Step 2 - Launching browser\n")
	c.browserFn(authURL)

	c.printdbg(fmt.Sprintf("Login: Step 3 - Wait for code on local server. Addr: %s\n", c.Conf.ListenAddr()))
	code, err := srv.waitForCode(ctx)
	if err != nil {
		c.printdbg(fmt.Sprintf("Failed to receive code on local server. Err: %v\n", err))
		return err
	}

	c.printdbg("Login: Step 4 - Exchange code for refresh and access tokens\n")
	tok, err := exchangeCode(ctx, c.Conf, code, verifier)
	if err != nil {
		c.printdbg(fmt.Sprintf("Failed to exchange code for refresh / access tokens. Err: %v\n", err))
		return err
	}

	clerkToken := ClerkToken{
		RefreshToken: tok.RefreshToken,
		AccessToken:  tok.AccessToken,
		Expiry:       tok.Expiry,
	}

	c.printdbg("Login: Step 5 - Save token using provided ClerkTokenMgr SaveTokenFn\n")
	err = c.tokenMgr.saveTokenFn(clerkToken)
	if err != nil {
		c.printdbg(fmt.Sprintf("Failed to save token. Err: %v\n", err))
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
	q.Set("redirect_uri", c.Conf.RedirectURL())
	q.Set("scope", "email offline_access profile")
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	authURL.RawQuery = q.Encode()
	return authURL.String(), nil
}

func (c *ClerkCli) getValidToken(
	ctx context.Context,
	forceRefresh bool,
) (*oauth2.Token, error) {
	// We lock this entire function in order to avoid refresh storms
	c.mu.Lock()
	defer c.mu.Unlock()
	ctok, err := c.tokenMgr.loadTokenFn()
	if err != nil {
		return nil, err
	}
	tok := ctok.toOauthToken()

	if !forceRefresh && tok.Expiry.After(time.Now().Add(30*time.Second)) {
		return tok, nil
	}

	if tok.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	cfg := c.Conf.toOAuth2Config()
	ts := cfg.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	newCtok := ClerkToken{
		AccessToken:  newTok.AccessToken,
		RefreshToken: newTok.RefreshToken,
		Expiry:       newTok.Expiry,
	}
	if err = c.tokenMgr.saveTokenFn(newCtok); err != nil {
		return nil, err
	}
	return newTok, nil
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

	cfg := conf.toOAuth2Config()
	// Scopes aren't used for token exchange
	cfg.Scopes = []string{}
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

func (c *ClerkCli) printdbg(msg string) {
	if !c.debug {
		return
	}
	fmt.Print(msg)
}
