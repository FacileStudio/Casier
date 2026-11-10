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
	store := newTestUserStore(t, db)

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
	store := newTestUserStore(t, db)

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

// TestUpsertFromOIDCRefusesARecycledAddress covers porte's second linking rule:
// an account that already holds one subject of the issuer is not handed to a
// second subject just because it presents the same verified address.
func TestUpsertFromOIDCRefusesARecycledAddress(t *testing.T) {
	db := resetTestDB(t)
	store := newTestUserStore(t, db)
	ctx := context.Background()

	owner, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer, Subject: "first-owner", Email: "shared@facile.studio", EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("first sign-in: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
		owner, testIssuer, "first-owner").Error; err != nil {
		t.Fatalf("seeding the identity porte writes after a sign-in: %v", err)
	}

	if _, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer + "/", Subject: "second-owner", Email: "shared@facile.studio", EmailVerified: true,
	}); err == nil {
		t.Fatal("a second subject of the same issuer linked an account that already has one")
	}

	again, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer, Subject: "first-owner", Email: "shared@facile.studio", EmailVerified: true,
	})
	if err != nil || again != owner {
		t.Fatalf("the real owner signing in again got (%d, %v), want (%d, nil)", again, err, owner)
	}
}

// TestUpsertFromOIDCStoresTheLocale checks the claim lands on the account and
// that a sign-in without one keeps what is stored.
func TestUpsertFromOIDCStoresTheLocale(t *testing.T) {
	db := resetTestDB(t)
	store := newTestUserStore(t, db)
	ctx := context.Background()

	id, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer, Subject: "polyglot", Email: "p@facile.studio", EmailVerified: true, Locale: "de",
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC: %v", err)
	}
	if _, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer, Subject: "polyglot", Email: "p@facile.studio", EmailVerified: true,
	}); err != nil {
		t.Fatalf("UpsertFromOIDC without a locale: %v", err)
	}
	var user schemas.User
	if err := db.First(&user, id).Error; err != nil {
		t.Fatalf("reading the user: %v", err)
	}
	if user.Locale != "de" {
		t.Errorf("locale = %q, want de", user.Locale)
	}
}

// TestUpsertFromOIDCRewritesTheEmailOnlyWhenVerified pins the suite rule
// porte/kit applies: an address the provider did not verify never replaces
// the stored one, because CASIER_ADMIN_EMAILS and invitations read it.
func TestUpsertFromOIDCRewritesTheEmailOnlyWhenVerified(t *testing.T) {
	db := resetTestDB(t)
	store := newTestUserStore(t, db)
	ctx := context.Background()
	claims := porte.Claims{Provider: testIssuer, Subject: "mover", Email: "old@facile.studio", EmailVerified: true}

	id, err := store.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatalf("first sign-in: %v", err)
	}
	exec(t, db, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'mover')`, id, testIssuer)
	for _, step := range []struct {
		email, want string
		verified    bool
	}{{"unverified@facile.studio", "old@facile.studio", false}, {"new@facile.studio", "new@facile.studio", true}} {
		claims.Email, claims.EmailVerified = step.email, step.verified
		if _, err := store.UpsertFromOIDC(ctx, claims); err != nil {
			t.Fatalf("sign-in with %s: %v", step.email, err)
		}
		var user schemas.User
		if err := db.First(&user, id).Error; err != nil {
			t.Fatalf("reading the user: %v", err)
		}
		if user.Email != step.want {
			t.Errorf("after a sign-in with %s (verified %v), email = %q, want %q", step.email, step.verified, user.Email, step.want)
		}
	}
}
