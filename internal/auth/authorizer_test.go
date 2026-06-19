package auth

import (
	"context"
	"errors"
	"testing"
)

func TestAdminTokenAuthorizerAllowsWritesWhenTokenIsNotConfigured(t *testing.T) {
	authorizer := NewAdminTokenAuthorizer("")

	if err := authorizer.Authorize(context.Background(), ""); err != nil {
		t.Fatalf("expected access without configured token, got %v", err)
	}
}

func TestAdminTokenAuthorizerRequiresMatchingBearerToken(t *testing.T) {
	authorizer := NewAdminTokenAuthorizer("secret")

	if err := authorizer.Authorize(context.Background(), "Bearer secret"); err != nil {
		t.Fatalf("expected matching token to be accepted, got %v", err)
	}
	if err := authorizer.Authorize(context.Background(), "Bearer wrong"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized for wrong token, got %v", err)
	}
}

func TestBearerTokenParsesBearerHeader(t *testing.T) {
	token, ok := BearerToken("Bearer abc.def.ghi")

	if !ok || token != "abc.def.ghi" {
		t.Fatalf("expected bearer token to be parsed, got token=%q ok=%t", token, ok)
	}
}
