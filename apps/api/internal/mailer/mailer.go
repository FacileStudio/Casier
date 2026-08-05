// Package mailer sends the handful of transactional messages Casier owes its
// users, starting with the password reset link. It is net/smtp and nothing
// else: one plaintext message over one connection is squarely inside what the
// standard library already does, and a secrets manager is the last place to
// take on an unaudited mail dependency.
package mailer

import "context"

type Message struct {
	To      string
	Subject string
	Body    string
}

// Mailer is the seam callers depend on, so a test can assert on what would have
// been sent without opening a socket.
//
// A nil Mailer is a supported state, not a bug: it means the deployment
// configured no relay. Callers must refuse the feature that needs mail rather
// than pretend a message went out, which is exactly the failure that left
// password reset silently inert before this package existed.
type Mailer interface {
	Send(ctx context.Context, message Message) error
}
