package auth

import (
	"strings"
	"testing"
)

func TestParseCLICallbackRejectsUnsafeInput(t *testing.T) {
	cases := []struct {
		name  string
		port  string
		state string
	}{
		{"empty port", "", "abc"},
		{"privileged port", "80", "abc"},
		{"port out of range", "70000", "abc"},
		{"negative port", "-1", "abc"},
		{"non numeric port", "12a4", "abc"},
		{"port with path traversal", "1234/../evil", "abc"},
		{"empty state", "1234", ""},
		{"state with host injection", "1234", "evil.com/x"},
		{"state with at sign", "1234", "a@evil.com"},
		{"state with crlf", "1234", "abc\r\nLocation: http://evil.com"},
		{"state too long", "1234", strings.Repeat("a", 65)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := parseCLICallback(tc.port, tc.state); ok {
				t.Fatalf("parseCLICallback(%q, %q) accepted unsafe input", tc.port, tc.state)
			}
		})
	}
}

func TestCLIRedirectURLStaysOnLoopback(t *testing.T) {
	cli, ok := parseCLICallback("52345", "s-t_a1")
	if !ok {
		t.Fatal("parseCLICallback rejected a valid callback")
	}

	got := cli.redirectURL("tok en/&x=1")
	if !strings.HasPrefix(got, "http://127.0.0.1:52345/callback?") {
		t.Fatalf("redirect escaped loopback: %s", got)
	}
	if strings.Contains(got, "tok en") || strings.Contains(got, "&x=1") {
		t.Fatalf("token was not escaped into the query: %s", got)
	}
	if !strings.Contains(got, "state=s-t_a1") {
		t.Fatalf("state missing from redirect: %s", got)
	}
}

func TestDecodeCLICallbackRoundTrip(t *testing.T) {
	cli, ok := parseCLICallback("49152", "abc123")
	if !ok {
		t.Fatal("parseCLICallback rejected a valid callback")
	}

	decoded, ok := decodeCLICallback(cli.encode())
	if !ok {
		t.Fatal("decodeCLICallback rejected its own encoding")
	}
	if decoded != cli {
		t.Fatalf("round trip changed the callback: %+v != %+v", decoded, cli)
	}

	if _, ok := decodeCLICallback("no-separator"); ok {
		t.Fatal("decodeCLICallback accepted a malformed cookie")
	}
	if _, ok := decodeCLICallback("80:abc"); ok {
		t.Fatal("decodeCLICallback skipped port validation")
	}
}
