package ai

import "errors"

// retryableError wraps an error returned by an AI provider when the failure is temporary
// and retrying the request may succeed (e.g. rate limit, server overload).
type retryableError struct {
	err error
}

func (e *retryableError) Error() string { return e.err.Error() }

func (e *retryableError) Unwrap() error { return e.err }

// IsRetryableError reports whether err or any error in its chain is a retryableError.
func IsRetryableError(err error) bool {
	var t *retryableError

	return errors.As(err, &t)
}

func newRetryableError(err error) *retryableError { return &retryableError{err: err} }
