package auth

import (
	"context"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type KeycloakConfig struct {
	IssuerURL string
	ClientID  string
}

type KeycloakAuthorizer struct {
	verifier *oidc.IDTokenVerifier
}

func NewKeycloakAuthorizer(ctx context.Context, cfg KeycloakConfig) (*KeycloakAuthorizer, error) {
	provider, err := oidc.NewProvider(ctx, strings.TrimRight(cfg.IssuerURL, "/"))
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID:          cfg.ClientID,
		SkipClientIDCheck: strings.TrimSpace(cfg.ClientID) == "",
	})

	return &KeycloakAuthorizer{verifier: verifier}, nil
}

func (a *KeycloakAuthorizer) Authorize(ctx context.Context, authorizationHeader string) error {
	token, ok := BearerToken(authorizationHeader)
	if !ok {
		return ErrUnauthorized
	}

	if _, err := a.verifier.Verify(ctx, token); err != nil {
		return ErrUnauthorized
	}

	return nil
}
