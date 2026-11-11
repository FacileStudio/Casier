package mcp

import (
	"testing"

	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// count runs a count query.
func (f *fixture) count(t testing.TB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := f.sqlDB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// call runs a tool and decodes its result into out, failing on an error result.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	mcptest.Decode(t, mcptest.Call(t, session, tool, args), out)
}

// refused runs a tool and returns its error text, failing when it succeeded.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must fail, got %s", tool, mcptest.Text(result))
	}
	return mcptest.Text(result)
}

// prod names the owner's prod environment.
func prod(extra map[string]any) map[string]any {
	args := map[string]any{"project": projectSlug, "environment": envSlug}
	for k, v := range extra {
		args[k] = v
	}
	return args
}
