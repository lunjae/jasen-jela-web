package services

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4/auth"
)

type FirebaseVerifier struct{ Client *firebase.Client }

func (v FirebaseVerifier) Verify(ctx context.Context, token string) (string, error) {
	t, e := v.Client.VerifyIDToken(ctx, token)
	if e != nil {
		return "", e
	}
	return t.UID, nil
}

type RejectVerifier struct{}

func (RejectVerifier) Verify(context.Context, string) (string, error) {
	return "", errors.New("token verification is not configured")
}
