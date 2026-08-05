package mailer

import (
	"context"
	"net/mail"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func fixedTime(t *testing.T) time.Time {
	t.Helper()
	stamp, err := time.Parse(time.RFC3339, "2026-08-05T10:30:00Z")
	if err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	return stamp
}

func mustAddress(t *testing.T, value string) *mail.Address {
	t.Helper()
	address, err := mail.ParseAddress(value)
	if err != nil {
		t.Fatalf("mail.ParseAddress(%q): %v", value, err)
	}
	return address
}

func TestNewSMTPValidatesItsConfiguration(t *testing.T) {
	cases := []struct {
		name    string
		config  SMTPConfig
		wantErr bool
	}{
		{
			name:   "host port and from are enough",
			config: SMTPConfig{Host: "smtp.internal", Port: 587, From: "casier@example.com"},
		},
		{
			name:   "a relay may take no credentials",
			config: SMTPConfig{Host: "smtp.internal", Port: 25, From: "Casier <casier@example.com>"},
		},
		{
			name:   "credentials are accepted in full",
			config: SMTPConfig{Host: "smtp.internal", Port: 587, Username: "casier", Password: "s3cret", From: "casier@example.com"},
		},
		{
			name:    "missing host",
			config:  SMTPConfig{Port: 587, From: "casier@example.com"},
			wantErr: true,
		},
		{
			name:    "blank host",
			config:  SMTPConfig{Host: "   ", Port: 587, From: "casier@example.com"},
			wantErr: true,
		},
		{
			name:    "port zero",
			config:  SMTPConfig{Host: "smtp.internal", From: "casier@example.com"},
			wantErr: true,
		},
		{
			name:    "port out of range",
			config:  SMTPConfig{Host: "smtp.internal", Port: 70000, From: "casier@example.com"},
			wantErr: true,
		},
		{
			name:    "missing from",
			config:  SMTPConfig{Host: "smtp.internal", Port: 587},
			wantErr: true,
		},
		{
			name:    "unparseable from",
			config:  SMTPConfig{Host: "smtp.internal", Port: 587, From: "not an address"},
			wantErr: true,
		},
		{
			name:    "username without a password",
			config:  SMTPConfig{Host: "smtp.internal", Port: 587, Username: "casier", From: "casier@example.com"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSMTP(tc.config)
			if tc.wantErr && err == nil {
				t.Fatal("newSMTP accepted an invalid configuration")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("newSMTP: %v", err)
			}
		})
	}
}

func TestNewWithoutAConfigurationYieldsNoMailer(t *testing.T) {
	got, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil): %v", err)
	}
	if got != nil {
		t.Fatal("New(nil) must return a nil Mailer, not an interface wrapping a nil pointer")
	}
}

func TestBuildMessage(t *testing.T) {
	now := fixedTime(t)

	cases := []struct {
		name        string
		from        string
		to          string
		message     Message
		wantErr     bool
		wantLines   []string
		wantAbsent  []string
		wantBodyEnd string
	}{
		{
			name:    "bare addresses",
			from:    "casier@example.com",
			to:      "user@example.org",
			message: Message{Subject: "Reset your Casier password", Body: "line one\nline two"},
			wantLines: []string{
				"From: <casier@example.com>",
				"To: <user@example.org>",
				"Subject: Reset your Casier password",
				"Date: Wed, 05 Aug 2026 10:30:00 +0000",
				"MIME-Version: 1.0",
				`Content-Type: text/plain; charset="utf-8"`,
			},
			wantBodyEnd: "line one\r\nline two",
		},
		{
			name:    "display names are preserved",
			from:    "Casier <casier@example.com>",
			to:      "Ada Lovelace <ada@example.org>",
			message: Message{Subject: "hello", Body: "body"},
			wantLines: []string{
				`From: "Casier" <casier@example.com>`,
				`To: "Ada Lovelace" <ada@example.org>`,
			},
		},
		{
			name:       "a non ascii subject is encoded",
			from:       "casier@example.com",
			to:         "user@example.org",
			message:    Message{Subject: "Réinitialisation", Body: "body"},
			wantLines:  []string{"Subject: =?utf-8?q?R=C3=A9initialisation?="},
			wantAbsent: []string{"Subject: Réinitialisation"},
		},
		{
			name:    "crlf in the body is not doubled",
			from:    "casier@example.com",
			to:      "user@example.org",
			message: Message{Subject: "hello", Body: "line one\r\nline two"},

			wantBodyEnd: "line one\r\nline two",
		},
		{
			name:    "a subject carrying a line break is refused",
			from:    "casier@example.com",
			to:      "user@example.org",
			message: Message{Subject: "hello\r\nBcc: attacker@example.net", Body: "body"},
			wantErr: true,
		},
		{
			name:    "a subject carrying a bare newline is refused",
			from:    "casier@example.com",
			to:      "user@example.org",
			message: Message{Subject: "hello\nBcc: attacker@example.net", Body: "body"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildMessage(mustAddress(t, tc.from), mustAddress(t, tc.to), tc.message, now)
			if tc.wantErr {
				if err == nil {
					t.Fatal("buildMessage accepted a header injection")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildMessage: %v", err)
			}

			rendered := string(got)
			headers, body, found := strings.Cut(rendered, "\r\n\r\n")
			if !found {
				t.Fatalf("no blank line separating headers from body: %q", rendered)
			}
			for _, line := range tc.wantLines {
				if !strings.Contains(headers+"\r\n", line+"\r\n") {
					t.Errorf("missing header %q in:\n%s", line, headers)
				}
			}
			for _, line := range tc.wantAbsent {
				if strings.Contains(headers, line) {
					t.Errorf("unexpected header %q in:\n%s", line, headers)
				}
			}
			if tc.wantBodyEnd != "" && body != tc.wantBodyEnd {
				t.Errorf("body = %q, want %q", body, tc.wantBodyEnd)
			}
			if strings.Contains(headers, "\n") && !strings.Contains(headers, "\r\n") {
				t.Error("headers must be CRLF terminated")
			}
		})
	}
}

type sentMail struct {
	addr string
	auth smtp.Auth
	from string
	to   []string
	msg  []byte
}

func recordingMailer(t *testing.T, config SMTPConfig, sent *sentMail) *smtpMailer {
	t.Helper()
	mailer, err := newSMTP(config)
	if err != nil {
		t.Fatalf("newSMTP: %v", err)
	}
	mailer.now = func() time.Time { return fixedTime(t) }
	mailer.send = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		*sent = sentMail{addr: addr, auth: auth, from: from, to: to, msg: msg}
		return nil
	}
	return mailer
}

func TestSendHandsTheRelayAnEnvelope(t *testing.T) {
	cases := []struct {
		name     string
		config   SMTPConfig
		to       string
		wantAddr string
		wantFrom string
		wantTo   []string
		wantAuth bool
	}{
		{
			name:     "an anonymous relay gets no auth",
			config:   SMTPConfig{Host: "smtp.internal", Port: 25, From: "Casier <casier@example.com>"},
			to:       "Ada Lovelace <ada@example.org>",
			wantAddr: "smtp.internal:25",
			wantFrom: "casier@example.com",
			wantTo:   []string{"ada@example.org"},
		},
		{
			name:     "credentials become plain auth",
			config:   SMTPConfig{Host: "smtp.internal", Port: 587, Username: "casier", Password: "s3cret", From: "casier@example.com"},
			to:       "user@example.org",
			wantAddr: "smtp.internal:587",
			wantFrom: "casier@example.com",
			wantTo:   []string{"user@example.org"},
			wantAuth: true,
		},
		{
			name:     "an ipv6 host is bracketed",
			config:   SMTPConfig{Host: "::1", Port: 2525, From: "casier@example.com"},
			to:       "user@example.org",
			wantAddr: "[::1]:2525",
			wantFrom: "casier@example.com",
			wantTo:   []string{"user@example.org"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sent sentMail
			mailer := recordingMailer(t, tc.config, &sent)

			if err := mailer.Send(context.Background(), Message{To: tc.to, Subject: "hello", Body: "body"}); err != nil {
				t.Fatalf("Send: %v", err)
			}
			if sent.addr != tc.wantAddr {
				t.Errorf("addr = %q, want %q", sent.addr, tc.wantAddr)
			}
			if sent.from != tc.wantFrom {
				t.Errorf("envelope from = %q, want %q", sent.from, tc.wantFrom)
			}
			if len(sent.to) != len(tc.wantTo) || sent.to[0] != tc.wantTo[0] {
				t.Errorf("envelope to = %v, want %v", sent.to, tc.wantTo)
			}
			if (sent.auth != nil) != tc.wantAuth {
				t.Errorf("auth present = %t, want %t", sent.auth != nil, tc.wantAuth)
			}
			if !strings.HasSuffix(string(sent.msg), "body") {
				t.Errorf("message body was not delivered: %q", string(sent.msg))
			}
		})
	}
}

func TestSendRejectsBadInputBeforeDialing(t *testing.T) {
	cases := []struct {
		name    string
		message Message
		ctx     func() context.Context
	}{
		{
			name:    "unparseable recipient",
			message: Message{To: "not an address", Subject: "hello", Body: "body"},
		},
		{
			name:    "recipient carrying a header break",
			message: Message{To: "user@example.org>\r\nBcc: attacker@example.net", Subject: "hello", Body: "body"},
		},
		{
			name:    "empty recipient",
			message: Message{Subject: "hello", Body: "body"},
		},
		{
			name:    "subject carrying a header break",
			message: Message{To: "user@example.org", Subject: "hello\r\nBcc: attacker@example.net", Body: "body"},
		},
		{
			name:    "cancelled context",
			message: Message{To: "user@example.org", Subject: "hello", Body: "body"},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dialed := false
			mailer, err := newSMTP(SMTPConfig{Host: "smtp.internal", Port: 587, From: "casier@example.com"})
			if err != nil {
				t.Fatalf("newSMTP: %v", err)
			}
			mailer.send = func(string, smtp.Auth, string, []string, []byte) error {
				dialed = true
				return nil
			}

			ctx := context.Background()
			if tc.ctx != nil {
				ctx = tc.ctx()
			}
			if err := mailer.Send(ctx, tc.message); err == nil {
				t.Fatal("Send accepted an invalid message")
			}
			if dialed {
				t.Fatal("Send reached the relay with an invalid message")
			}
		})
	}
}
