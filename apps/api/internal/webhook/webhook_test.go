package webhook

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Casier/apps/api/schemas"
)

type delivery struct {
	signature string
	body      []byte
}

func newTestDispatcher(webhooks []schemas.Webhook) *Dispatcher {
	return &Dispatcher{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		client: &http.Client{Timeout: deliveryTimeout},
		load: func(ctx context.Context, projectID int64) ([]schemas.Webhook, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return webhooks, nil
		},
	}
}

func TestFireDeliversAfterCallerContextIsCancelled(t *testing.T) {
	received := make(chan delivery, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received <- delivery{signature: r.Header.Get("X-Casier-Signature"), body: body}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := newTestDispatcher([]schemas.Webhook{
		{ID: 1, ProjectID: 7, URL: server.URL, Secret: "s3cr3t", Events: "*", Active: true},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := Event{Type: "secret.updated", ProjectID: 7, Environment: "prod", Key: "API_KEY"}
	dispatcher.Fire(ctx, event)

	select {
	case got := <-received:
		want := Signature("s3cr3t", got.body)
		if got.signature != want {
			t.Fatalf("signature = %q, want %q", got.signature, want)
		}
		if got.signature == "sha256=" {
			t.Fatal("signature header is empty")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("webhook was not delivered after the caller context was cancelled")
	}
}

func TestSignatureIsHMACSHA256(t *testing.T) {
	got := Signature("key", []byte("payload"))
	want := "sha256=5d98b45c90a207fa998ce639fea6f02ecc8cc3f36fef81d694fb856b4d0a28ca"
	if got != want {
		t.Fatalf("Signature = %q, want %q", got, want)
	}
}

func TestMatchesEvent(t *testing.T) {
	cases := []struct {
		events    string
		eventType string
		want      bool
	}{
		{"", "secret.created", true},
		{"*", "secret.created", true},
		{"secret.created, secret.deleted", "secret.deleted", true},
		{"secret.created", "secret.deleted", false},
	}

	for _, c := range cases {
		if got := matchesEvent(c.events, c.eventType); got != c.want {
			t.Errorf("matchesEvent(%q, %q) = %v, want %v", c.events, c.eventType, got, c.want)
		}
	}
}
