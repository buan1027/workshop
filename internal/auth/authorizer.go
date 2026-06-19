package auth

import (
	"context"
	"errors"
	"strings"
)

var ErrUnauthorized = errors.New("unauthorized")

type Authorizer interface {
	Authorize(ctx context.Context, authorizationHeader string) error
}

type NoopAuthorizer struct{}

func (NoopAuthorizer) Authorize(context.Context, string) error {
	return nil
}

type AdminTokenAuthorizer struct {
	token string
}

func NewAdminTokenAuthorizer(token string) AdminTokenAuthorizer {
	return AdminTokenAuthorizer{token: strings.TrimSpace(token)}
}

func (a AdminTokenAuthorizer) Authorize(_ context.Context, authorizationHeader string) error {
	if a.token == "" {
		return nil
	}
	if authorizationHeader == "Bearer "+a.token {
		return nil
	}
	return ErrUnauthorized
}

func BearerToken(authorizationHeader string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(authorizationHeader), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}
	return strings.TrimSpace(token), true
}
