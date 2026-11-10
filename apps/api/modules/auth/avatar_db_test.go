package auth

import (
	"context"
	"testing"

	"github.com/FacileStudio/Casier/apps/api/schemas"
	"github.com/FacileStudio/porte"

	"gorm.io/gorm"
)

const testIssuer = "https://sso.facile.studio"

// seedAvatarUsers creates one user per way of arriving: signed in with a real
// picture, signed in through another provider with a data: placeholder, and
// signed in before Casier stored the claim. It returns their ids in that order.
func seedAvatarUsers(t *testing.T, db *gorm.DB, store *UserStore) (int64, int64, int64) {
	t.Helper()
	ctx := context.Background()
	signedIn, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: testIssuer, Subject: "sub-a", Email: "a@facile.studio", EmailVerified: true,
		Picture: "https://cdn.example/a.png",
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC: %v", err)
	}
	placeholder, err := store.UpsertFromOIDC(ctx, porte.Claims{
		Provider: "https://other.example", Subject: "sub-b", Email: "b@facile.studio", EmailVerified: true,
		Picture: "data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=",
	})
	if err != nil {
		t.Fatalf("UpsertFromOIDC: %v", err)
	}
	stale := schemas.User{Email: "c@facile.studio"}
	if err := db.Create(&stale).Error; err != nil {
		t.Fatalf("seeding: %v", err)
	}
	for _, row := range []map[string]any{
		{"user_id": signedIn, "provider": testIssuer, "subject": "sub-a"},
		{"user_id": placeholder, "provider": "https://other.example", "subject": "sub-b"},
		{"user_id": stale.ID, "provider": testIssuer, "subject": "sub-c"},
	} {
		if err := db.Table("porte_identities").Create(row).Error; err != nil {
			t.Fatalf("seeding identity: %v", err)
		}
	}
	return signedIn, placeholder, stale.ID
}

// TestAvatarURLFollowsTheProvider pins both ways a user gets a picture URL:
// a sign-in writes the https claim (and refuses a data: placeholder), and the
// boot backfill derives the same URL for someone who has not signed in since,
// without touching a URL a sign-in already wrote.
func TestAvatarURLFollowsTheProvider(t *testing.T) {
	db := resetTestDB(t)
	store := newTestUserStore(t, db)
	signedIn, placeholder, stale := seedAvatarUsers(t, db, store)

	for range 2 {
		if _, err := store.BackfillPictures(context.Background(), testIssuer+"/"); err != nil {
			t.Fatalf("BackfillPictures: %v", err)
		}
	}

	want := map[int64]string{
		signedIn:    "https://cdn.example/a.png",
		placeholder: "",
		stale:       testIssuer + "/avatars/sub-c",
	}
	for id, url := range want {
		var user schemas.User
		if err := db.First(&user, id).Error; err != nil {
			t.Fatalf("reading user %d: %v", id, err)
		}
		if user.AvatarURL != url {
			t.Errorf("user %s avatar_url = %q, want %q", user.Email, user.AvatarURL, url)
		}
	}
}
