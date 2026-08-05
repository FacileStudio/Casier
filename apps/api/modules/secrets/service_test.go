package secrets

import (
	"strings"
	"testing"
)

func TestEnvValueRoundTrip(t *testing.T) {
	values := map[string]string{
		"plain":              "hunter2",
		"empty":              "",
		"spaces":             "  padded  ",
		"quote":              `say "hi"`,
		"backslash":          `C:\Users\casier`,
		"escape sequence":    `not\na newline`,
		"multiline":          "-----BEGIN KEY-----\nline two\nline three\n",
		"carriage return":    "crlf\r\nvalue",
		"tab":                "left\tright",
		"comment marker":     "value # not a comment",
		"single quotes":      "it's fine",
		"everything at once": "a\"b\\c\nd#e\tf 'g'",
	}

	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			encoded := encodeEnvValue(value)
			if strings.ContainsAny(encoded, "\n\r") {
				t.Fatalf("encodeEnvValue(%q) = %q, which breaks the one-secret-per-line format", value, encoded)
			}

			entries, skipped := parseEnvFile("KEY=" + encoded + "\n")
			if skipped != 0 {
				t.Fatalf("parseEnvFile skipped %d lines for %q", skipped, encoded)
			}
			if len(entries) != 1 {
				t.Fatalf("parseEnvFile(%q) returned %d entries, want 1", encoded, len(entries))
			}
			if entries[0].key != "KEY" {
				t.Fatalf("parsed key = %q, want %q", entries[0].key, "KEY")
			}
			if entries[0].value != value {
				t.Fatalf("round trip of %q through %q gave %q", value, encoded, entries[0].value)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	longKey := strings.Repeat("K", maxSecretKeyLength+1)
	content := strings.Join([]string{
		"# a comment",
		"",
		"  ",
		"API_KEY=abc",
		"NO_EQUALS_SIGN",
		"=orphan value",
		longKey + "=too long",
		"  SPACED  =  trimmed  ",
		"LEGACY='single quoted'",
		"EMPTY=",
	}, "\n")

	entries, skipped := parseEnvFile(content)
	if skipped != 3 {
		t.Fatalf("skipped = %d, want 3", skipped)
	}

	want := []envEntry{
		{key: "API_KEY", value: "abc"},
		{key: "SPACED", value: "trimmed"},
		{key: "LEGACY", value: "single quoted"},
		{key: "EMPTY", value: ""},
	}
	if len(entries) != len(want) {
		t.Fatalf("parsed %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, entry := range entries {
		if entry != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, entry, want[i])
		}
	}
}

func TestTagFilterPattern(t *testing.T) {
	cases := []struct {
		name string
		tag  string
		want string
	}{
		{"plain tag is padded with commas", "prod", "%,prod,%"},
		{"percent loses its wildcard meaning", "pr%d", `%,pr\%d,%`},
		{"underscore loses its wildcard meaning", "pr_d", `%,pr\_d,%`},
		{"backslash is escaped first", `a\b`, `%,a\\b,%`},
		{"bare wildcard cannot match everything", "%", `%,\%,%`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := tagFilterPattern(testCase.tag); got != testCase.want {
				t.Fatalf("tagFilterPattern(%q) = %q, want %q", testCase.tag, got, testCase.want)
			}
		})
	}
}

// TestTagFilterMatching mirrors what Postgres does with the pattern so the
// comma padding is checked against real tag lists, not just its own spelling.
func TestTagFilterMatching(t *testing.T) {
	matches := func(tags, filter string) bool {
		pattern := tagFilterPattern(filter)
		unescaped := strings.NewReplacer(`\%`, "%", `\_`, "_", `\\`, `\`).Replace(pattern)
		needle := strings.TrimSuffix(strings.TrimPrefix(unescaped, "%"), "%")
		return strings.Contains(","+tags+",", needle)
	}

	cases := []struct {
		tags   string
		filter string
		want   bool
	}{
		{"prod,db", "prod", true},
		{"db,prod", "prod", true},
		{"db,prod,cache", "prod", true},
		{"prod", "prod", true},
		{"not-production", "prod", false},
		{"production", "prod", false},
		{"staging", "prod", false},
	}

	for _, testCase := range cases {
		if got := matches(testCase.tags, testCase.filter); got != testCase.want {
			t.Fatalf("tags %q filtered by %q = %t, want %t", testCase.tags, testCase.filter, got, testCase.want)
		}
	}
}

func TestACLDecisions(t *testing.T) {
	unrestricted := secretGrant{}
	readOnly := secretGrant{restricted: true, canRead: true}
	writable := secretGrant{restricted: true, canRead: true, canWrite: true}
	denied := secretGrant{restricted: true}

	cases := []struct {
		name      string
		role      string
		grant     secretGrant
		wantRead  bool
		wantWrite bool
	}{
		{"secret without ACLs follows the project role", "member", unrestricted, true, true},
		{"owner bypasses a denying ACL", roleOwner, denied, true, true},
		{"owner bypasses a read-only ACL", roleOwner, readOnly, true, true},
		{"admin is bound by ACLs like anyone else", "admin", denied, false, false},
		{"granted read cannot write", "member", readOnly, true, false},
		{"granted write can do both", "member", writable, true, true},
		{"restricted secret without a grant is invisible", "member", denied, false, false},
		{"viewer without ACLs still reads", "viewer", unrestricted, true, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := aclAllowsRead(testCase.role, testCase.grant); got != testCase.wantRead {
				t.Errorf("aclAllowsRead(%q, %+v) = %t, want %t", testCase.role, testCase.grant, got, testCase.wantRead)
			}
			if got := aclAllowsWrite(testCase.role, testCase.grant); got != testCase.wantWrite {
				t.Errorf("aclAllowsWrite(%q, %+v) = %t, want %t", testCase.role, testCase.grant, got, testCase.wantWrite)
			}
		})
	}
}
