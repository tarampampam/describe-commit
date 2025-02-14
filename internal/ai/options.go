package ai

type (
	// options is a set of options that can be applied to the AI provider.
	options struct {
		ShortMessageOnly bool
		EnableEmoji      bool
		MaxOutputTokens  int64
	}

	// Option is a function that modifies the options.
	Option func(*options)
)

// Apply applies the given options.
func (o options) Apply(opts ...Option) options {
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

// WithShortMessageOnly forces the provider to return only the short commit message (usually the first line).
func WithShortMessageOnly(on bool) Option { return func(o *options) { o.ShortMessageOnly = on } }

// WithEmoji enables or disables emoji in the commit message.
func WithEmoji(on bool) Option { return func(o *options) { o.EnableEmoji = on } }

// WithMaxOutputTokens sets the maximum number of tokens in the output.
func WithMaxOutputTokens(max int64) Option { return func(o *options) { o.MaxOutputTokens = max } }
