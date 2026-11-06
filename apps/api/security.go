package main

import (
	"log/slog"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
)

const iconifyScript = "https://cdn.jsdelivr.net/npm/iconify-icon@2.2.0/dist/iconify-icon.min.js"

// securityConfig extends tronc's default security headers with what the built
// client needs: the inline scripts adapter-static writes into its pages, the
// iconify-icon bundle app.html loads from jsDelivr at one exact URL, and the
// two APIs the client calls (Ardoise's plan catalog and the icon API). A client
// directory that cannot be hashed is logged and leaves the default script
// policy, which only reports.
func securityConfig(clientDir string, appLogger *slog.Logger) troncmiddleware.SecurityConfig {
	scripts := []string{iconifyScript}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			appLogger.Warn("failed to hash the client's inline scripts", slog.Any("error", err))
		} else {
			scripts = append(scripts, hashes...)
		}
	}
	return troncmiddleware.SecurityConfig{CSPSources: map[string][]string{
		"connect-src": {"https://api-ardoise.facile.studio", "https://api.iconify.design"},
		"script-src":  scripts,
	}}
}
