package auth

import (
	"encoding/json"
	"testing"
)

func TestKeycloakClaimsMatchClientByAuthorizedParty(t *testing.T) {
	claims := keycloakClaims{AuthorizedParty: "workshop-server"}

	if !claims.matchesClient("workshop-server") {
		t.Fatal("expected azp to match client id")
	}
}

func TestKeycloakClaimsMatchClientByAudience(t *testing.T) {
	claims := keycloakClaims{Audience: audience{"account", "workshop-server"}}

	if !claims.matchesClient("workshop-server") {
		t.Fatal("expected audience to match client id")
	}
}

func TestKeycloakClaimsRejectDifferentClient(t *testing.T) {
	claims := keycloakClaims{Audience: audience{"account"}, AuthorizedParty: "other-client"}

	if claims.matchesClient("workshop-server") {
		t.Fatal("expected different client to be rejected")
	}
}

func TestAudienceAcceptsStringOrArray(t *testing.T) {
	var single struct {
		Audience audience `json:"aud"`
	}
	if err := json.Unmarshal([]byte(`{"aud":"account"}`), &single); err != nil {
		t.Fatalf("unmarshal single audience: %v", err)
	}
	if len(single.Audience) != 1 || single.Audience[0] != "account" {
		t.Fatalf("unexpected single audience: %+v", single.Audience)
	}

	var multiple struct {
		Audience audience `json:"aud"`
	}
	if err := json.Unmarshal([]byte(`{"aud":["account","workshop-server"]}`), &multiple); err != nil {
		t.Fatalf("unmarshal multiple audience: %v", err)
	}
	if len(multiple.Audience) != 2 || multiple.Audience[1] != "workshop-server" {
		t.Fatalf("unexpected multiple audience: %+v", multiple.Audience)
	}
}
