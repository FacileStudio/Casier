package mcp

import (
	stderrors "errors"
	"testing"
	"time"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// TestAFailedReadKeepsTheConfirmation checks collect spends a confirmation only after
// its read succeeded, so a failed read can be retried.
func TestAFailedReadKeepsTheConfirmation(t *testing.T) {
	f := newFixture(t)
	id := confirmedReveal(t, f, "DATABASE_URL")
	who := f.askedBy(t, id)
	failed := stderrors.New("read failed")
	if err := collect(t.Context(), f.sqlDB, who, id, func(mcpkit.Action) error { return failed }); !stderrors.Is(err, failed) {
		t.Fatalf("want the read's error, got %v", err)
	}
	if got := f.outcomeOf(t, id); got != mcpkit.OutcomeDone {
		t.Fatalf("a failed read must not spend the confirmation, got %q", got)
	}
	var out revealOutput
	call(t, f.session(t, ownerSubject), toolReveal, collectArgs(id), &out)
	if out.Status != statusRevealed || len(out.Secrets) != 1 {
		t.Fatalf("want the value after the failed read, got %+v", out)
	}
}

// TestTwoCollectionsHandOverOnce checks a collection racing one in progress waits for
// it and is then refused, so values leave Casier once per confirmation.
func TestTwoCollectionsHandOverOnce(t *testing.T) {
	f := newFixture(t)
	id := confirmedReveal(t, f, "DATABASE_URL")
	who := f.askedBy(t, id)
	ctx := t.Context()
	entered, release := make(chan struct{}), make(chan struct{})
	first, second := make(chan error, 1), make(chan error, 1)
	go func() {
		first <- collect(ctx, f.sqlDB, who, id, func(mcpkit.Action) error { close(entered); <-release; return nil })
	}()
	<-entered
	go func() { second <- collect(ctx, f.sqlDB, who, id, func(mcpkit.Action) error { return nil }) }()
	select {
	case err := <-second:
		t.Fatalf("the second collection must wait for the first, got %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	if err1, err2 := <-first, <-second; err1 != nil || !stderrors.Is(err2, ErrNotConfirmed) {
		t.Fatalf("want the first collection to succeed and the second refused, got %v and %v", err1, err2)
	}
}
