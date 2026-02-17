package clerkcligo

import "fmt"

type ClerkToken struct {
	RefreshToken string
	AccessToken  string
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
