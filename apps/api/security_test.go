package main

import (
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestSecurityConfigAllowsNoVisionScript(t *testing.T) {
	cfg := securityConfig(t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, source := range cfg.CSPSources["script-src"] {
		if strings.Contains(source, "vision.facile.studio") {
			t.Errorf("script-src allows the Vision origin %s", source)
		}
	}
}
