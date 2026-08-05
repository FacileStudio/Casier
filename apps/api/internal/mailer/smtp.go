package mailer

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

const DefaultPort = 587

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type smtpMailer struct {
	config SMTPConfig
	from   *mail.Address
	// send is smtp.SendMail in production and a recorder in tests. net/smtp
	// offers no way to observe a message short of standing up a real listener,
	// and the interesting part here is what we hand it, not the wire protocol.
	send func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
	now  func() time.Time
}

// New turns a loaded configuration into a Mailer. A nil config yields a nil
// Mailer and no error: a deployment without a relay is a valid deployment, and
// the caller decides which features that takes offline.
func New(config *SMTPConfig) (Mailer, error) {
	if config == nil {
		return nil, nil
	}
	return newSMTP(*config)
}

func newSMTP(config SMTPConfig) (*smtpMailer, error) {
	if strings.TrimSpace(config.Host) == "" {
		return nil, fmt.Errorf("mailer: SMTP host is required")
	}
	if config.Port < 1 || config.Port > 65535 {
		return nil, fmt.Errorf("mailer: SMTP port %d is not a valid TCP port", config.Port)
	}
	from, err := mail.ParseAddress(config.From)
	if err != nil {
		return nil, fmt.Errorf("mailer: invalid from address %q: %w", config.From, err)
	}
	if config.Username != "" && config.Password == "" {
		return nil, fmt.Errorf("mailer: SMTP password is required when a username is set")
	}
	return &smtpMailer{config: config, from: from, send: smtp.SendMail, now: time.Now}, nil
}

func (m *smtpMailer) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	recipient, err := mail.ParseAddress(message.To)
	if err != nil {
		return fmt.Errorf("mailer: invalid recipient: %w", err)
	}

	body, err := buildMessage(m.from, recipient, message, m.now())
	if err != nil {
		return err
	}

	// Credentials stay optional: an internal relay on a trusted network often
	// has none, and smtp.SendMail skips AUTH entirely when handed a nil Auth.
	var auth smtp.Auth
	if m.config.Username != "" {
		auth = smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	}

	addr := net.JoinHostPort(m.config.Host, strconv.Itoa(m.config.Port))
	if err := m.send(addr, auth, m.from.Address, []string{recipient.Address}, body); err != nil {
		return fmt.Errorf("mailer: send failed: %w", err)
	}
	return nil
}

// buildMessage renders RFC 5322 headers plus a plaintext body. Addresses are
// re-rendered from their parsed form and the subject is refused outright if it
// carries a bare CR or LF, so nothing a caller passes can smuggle an extra
// header past the blank line.
func buildMessage(from *mail.Address, to *mail.Address, message Message, now time.Time) ([]byte, error) {
	if strings.ContainsAny(message.Subject, "\r\n") {
		return nil, fmt.Errorf("mailer: subject must not contain a line break")
	}

	var builder strings.Builder
	writeHeader(&builder, "From", from.String())
	writeHeader(&builder, "To", to.String())
	writeHeader(&builder, "Subject", mime.QEncoding.Encode("utf-8", message.Subject))
	writeHeader(&builder, "Date", now.Format(time.RFC1123Z))
	writeHeader(&builder, "MIME-Version", "1.0")
	writeHeader(&builder, "Content-Type", `text/plain; charset="utf-8"`)
	builder.WriteString("\r\n")
	builder.WriteString(normalizeNewlines(message.Body))

	return []byte(builder.String()), nil
}

func writeHeader(builder *strings.Builder, name string, value string) {
	builder.WriteString(name)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\r\n")
}

func normalizeNewlines(body string) string {
	return strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
}
