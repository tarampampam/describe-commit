package providers

import "context"

type (
	options struct {
		ShortMessageOnly bool
	}
	Option func(*options)
)

func (o options) Apply(opts ...Option) options {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithShortMessageOnly forces the provider to return only the short commit message (usually the first line).
func WithShortMessageOnly(on bool) Option { return func(o *options) { o.ShortMessageOnly = on } }

type Provider interface {
	// Query the remote provider for the given string.
	Query(context.Context, string, ...Option) (string, error)
}
