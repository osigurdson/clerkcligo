package clerkcligo

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

type ClerkToken struct {
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	Expiry       time.Time `json:"expiry"`
}

type ClerkTokenMgr struct {
	saveTokenFn func(ClerkToken) error
	loadTokenFn func() (ClerkToken, error)
}

func NewClerkTokenMgr(
	saveTokenFn func(ClerkToken) error,
	loadTokenFn func() (ClerkToken, error),
) (ClerkTokenMgr, error) {
	if saveTokenFn == nil {
		return ClerkTokenMgr{}, fmt.Errorf("saveTokenFn required")
	}

	if loadTokenFn == nil {
		return ClerkTokenMgr{}, fmt.Errorf("loadTokenFn required")
	}

	return ClerkTokenMgr{
		saveTokenFn: saveTokenFn,
		loadTokenFn: loadTokenFn,
	}, nil
}

func (c ClerkToken) toOauthToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		Expiry:       c.Expiry,
	}
}
