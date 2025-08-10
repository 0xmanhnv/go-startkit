package ports

import "context"

// EmailSender abstracts email sending so application layer does not depend on infrastructure.
type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SendEmailFunc adapts a function to the EmailSender interface.
type SendEmailFunc func(ctx context.Context, to, subject, body string) error

func (f SendEmailFunc) Send(ctx context.Context, to, subject, body string) error {
	return f(ctx, to, subject, body)
}
