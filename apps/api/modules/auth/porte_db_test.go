package auth

import (
	"context"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"
)

// TestUpsertFromOIDCLinksAVerifiedAddress covers the registre cutover: both
// halves of the (provider, subject) key change, so every existing user arrives
// looking brand new. The verified-email fallback is what hands them their own
// account back instead of a fresh empty one.
//
// The guard here was inverted before 2026-08-25 — it consulted the email only
// when the address was NOT verified — so this path fell through to Create and
// died on the unique index.
func TestUpsertFromOIDCLinksAVerifiedAddress(t *testing.T) {
	db := resetTestDB(t)
	store := NewUserStore(db)

	existing := schemas.User{Email: "camille@facile.studio", Name: "Camille"}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("seeding the existing account: %v", err)
	}

	id, err := store.UpsertFromOIDC(context.Background(), porte.Claims{
		Provider:      "https://sso.facile.studio",
		Subject:       "d0777c5a-b831-4f5c-9c5f-098c554e08fb",
		Email:         "Camille@Facile.Studio",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("a verified address must link the existing account, got: %v", err)
	}
	if id != existing.ID {
		t.Fatalf("linked user id = %d, want the existing %d — a new row here orphans every secret they own", id, existing.ID)
	}

	var count int64
	if err := db.Model(&schemas.User{}).Count(&count).Error; err != nil {
		t.Fatalf("counting users: %v", err)
	}
	if count != 1 {
		t.Errorf("user count = %d, want 1: the account forked instead of linking", count)
	}
}

// TestUpsertFromOIDCRefusesAnUnverifiedAddress is the other half of the guard.
// Matching on an address the provider did not verify is an account takeover
// primitive, so it must stay refused.
func TestUpsertFromOIDCRefusesAnUnverifiedAddress(t *testing.T) {
	db := resetTestDB(t)
	store := NewUserStore(db)

	existing := schemas.User{Email: "noah@facile.studio", Name: "Noah"}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("seeding the existing account: %v", err)
	}

	if _, err := store.UpsertFromOIDC(context.Background(), porte.Claims{
		Provider:      "https://sso.facile.studio",
		Subject:       "attacker-subject",
		Email:         "noah@facile.studio",
		EmailVerified: false,
	}); err == nil {
		t.Fatal("an unverified address matched an existing account; that is account takeover")
	}
}
