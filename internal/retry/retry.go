package retry

import (
	"context"
	"time"
)

// Option configures the Do function.
type Option func(*options)

type options struct {
	maxAttempts uint
	delay       time.Duration
}

// WithMaxAttempts sets the maximum number of times fn is called. Zero means unlimited retries.
func WithMaxAttempts(n uint) Option { return func(o *options) { o.maxAttempts = n } }

// WithDelay sets the pause between consecutive attempts. Zero (the default) means no delay.
func WithDelay(d time.Duration) Option { return func(o *options) { o.delay = d } }

// Do calls fn until it returns a nil error or signals stop (first return value is true). When the stop signal
// is true, Do returns fn's error immediately - even if it is non-nil. If the context is canceled, Do returns
// ctx.Err() immediately.
//
// By default, Do retries indefinitely with no delay between attempts. Use WithMaxAttempts to cap the number of
// attempts and WithDelay to insert a pause between them.
func Do(
	ctx context.Context,
	fn func(ctx context.Context, attempt uint) (bool, error),
	opts ...Option,
) error {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	var lastErr error

	for attempt := uint(0); o.maxAttempts == 0 || attempt < o.maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		var stop bool

		stop, lastErr = fn(ctx, attempt)
		if lastErr == nil || stop {
			return lastErr
		}

		if o.maxAttempts == 0 || attempt+1 < o.maxAttempts {
			if err := waitForNext(ctx, o.delay); err != nil {
				return err
			}
		}
	}

	return lastErr
}

// waitForNext pauses for d before the next attempt. With zero delay it only checks whether ctx is already
// canceled. With a positive delay it blocks until d elapses or ctx is canceled, whichever comes first.
func waitForNext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	t := time.NewTimer(d)

	select {
	case <-ctx.Done():
		if !t.Stop() {
			<-t.C
		}

		return ctx.Err()
	case <-t.C:
		return nil
	}
}
