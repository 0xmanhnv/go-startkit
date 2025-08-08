package ports

import "context"

// SMSSender abstracts SMS sending so application layer does not depend on infrastructure.
type SMSSender interface {
	Send(ctx context.Context, to, message string) error
}

// SendSMSFunc adapts a function to the SMSSender interface.
type SendSMSFunc func(ctx context.Context, to, message string) error

func (f SendSMSFunc) Send(ctx context.Context, to, message string) error { return f(ctx, to, message) }
