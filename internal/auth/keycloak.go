package auth

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type KeycloakConfig struct {
	IssuerURL string
	ClientID  string
}

type KeycloakAuthorizer struct {
	verifier *oidc.IDTokenVerifier
	clientID string
}

func NewKeycloakAuthorizer(ctx context.Context, cfg KeycloakConfig) (*KeycloakAuthorizer, error) {
	provider, err := oidc.NewProvider(ctx, strings.TrimRight(cfg.IssuerURL, "/"))
	if err != nil {
		return nil, err
	}

	clientID := strings.TrimSpace(cfg.ClientID)
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	return &KeycloakAuthorizer{verifier: verifier, clientID: clientID}, nil
}

func (a *KeycloakAuthorizer) Authorize(ctx context.Context, authorizationHeader string) error {
	token, ok := BearerToken(authorizationHeader)
	if !ok {
		return ErrUnauthorized
	}

	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return ErrUnauthorized
	}

	var claims keycloakClaims
	if err := idToken.Claims(&claims); err != nil {
		return ErrUnauthorized
	}
	if !claims.matchesClient(a.clientID) {
		return ErrUnauthorized
	}

	return nil
}

type keycloakClaims struct {
	Audience        audience `json:"aud"`
	AuthorizedParty string   `json:"azp"`
}

func (c keycloakClaims) matchesClient(clientID string) bool {
	if clientID == "" {
		return true
	}
	if c.AuthorizedParty == clientID {
		return true
	}
	for _, value := range c.Audience {
		if value == clientID {
			return true
		}
	}
	return false
}

type audience []string

func (a *audience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}
	*a = multiple
	return nil
}
