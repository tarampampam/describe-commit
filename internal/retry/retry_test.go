package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gh.tarampamp.am/describe-commit/internal/retry"
)

func TestDo(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	for name, tc := range map[string]struct {
		giveFn   func(context.Context, uint) (bool, error)
		giveOpts []retry.Option

		wantCalls int
		checkErr  func(*testing.T, error) // nil means no error expected
	}{
		"success on first attempt": {
			giveFn:    func(_ context.Context, _ uint) (bool, error) { return false, nil },
			wantCalls: 1,
		},
		"success after retries": {
			giveFn: func() func(context.Context, uint) (bool, error) {
				var n int

				return func(_ context.Context, _ uint) (bool, error) {
					n++
					if n < 3 {
						return false, errFoo
					}

					return false, nil
				}
			}(),
			giveOpts:  []retry.Option{retry.WithMaxAttempts(5)},
			wantCalls: 3,
		},
		"max attempts exhausted returns last error": {
			giveFn:    func(_ context.Context, _ uint) (bool, error) { return false, errFoo },
			giveOpts:  []retry.Option{retry.WithMaxAttempts(3)},
			wantCalls: 3,
			checkErr:  func(t *testing.T, err error) { assertErrorIs(t, err, errFoo) },
		},
		"stop signal with error returns that error": {
			giveFn:    func(_ context.Context, _ uint) (bool, error) { return true, errFoo },
			wantCalls: 1,
			checkErr:  func(t *testing.T, err error) { assertErrorIs(t, err, errFoo) },
		},
		"stop signal with nil error": {
			giveFn:    func(_ context.Context, _ uint) (bool, error) { return true, nil },
			wantCalls: 1,
		},
		"attempt counter starts at zero and increments": {
			giveFn: func(_ context.Context, attempt uint) (bool, error) {
				if attempt == 2 {
					return false, nil
				}

				return false, errFoo
			},
			giveOpts:  []retry.Option{retry.WithMaxAttempts(5)},
			wantCalls: 3,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var calls int

			err := retry.Do(t.Context(), func(ctx context.Context, attempt uint) (bool, error) {
				calls++

				return tc.giveFn(ctx, attempt)
			}, tc.giveOpts...)

			if tc.checkErr != nil {
				tc.checkErr(t, err)
			} else {
				assertNoError(t, err)
			}

			assertEqual(t, calls, tc.wantCalls)
		})
	}

	t.Run("context already cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var calls int

		err := retry.Do(ctx, func(_ context.Context, _ uint) (bool, error) {
			calls++

			return false, errFoo
		})

		assertErrorIs(t, err, context.Canceled)
		assertEqual(t, calls, 0)
	})

	t.Run("context cancelled between attempts without delay", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		var calls int

		err := retry.Do(ctx, func(_ context.Context, _ uint) (bool, error) {
			calls++

			cancel()

			return false, errFoo
		}, retry.WithMaxAttempts(5))

		assertErrorIs(t, err, context.Canceled)
		assertEqual(t, calls, 1)
	})

	t.Run("context cancelled during delay", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		var calls int

		err := retry.Do(ctx, func(_ context.Context, _ uint) (bool, error) {
			calls++

			cancel()

			return false, errFoo
		}, retry.WithDelay(time.Hour), retry.WithMaxAttempts(5))

		assertErrorIs(t, err, context.Canceled)
		assertEqual(t, calls, 1)
	})

	// regression: with maxAttempts=0 (infinite), delay must still be applied between attempts;
	// a 1s delay inside a 50ms window means fn can be called at most once before the deadline fires.
	t.Run("delay applied between infinite retries", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		var calls int

		err := retry.Do(ctx, func(_ context.Context, _ uint) (bool, error) {
			calls++

			return false, errFoo
		}, retry.WithDelay(time.Second))

		assertErrorIs(t, err, context.DeadlineExceeded)

		if calls > 2 {
			t.Errorf("delay not applied: fn called %d times (expected ≤ 2 with 1s delay in 50ms window)", calls)
		}
	})
}

// assertNoError fails the test if err is not nil, indicating an unexpected error occurred.
func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// assertErrorIs fails the test if err does not match target via errors.Is.
func assertErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("got error %v, want errors.Is match for %v", err, target)
	}
}

// assertEqual fails the test if got and want are not equal.
func assertEqual[T comparable](t *testing.T, got, want T, msgPrefix ...string) {
	t.Helper()

	var p string

	if len(msgPrefix) > 0 {
		p = msgPrefix[0] + ": "
	}

	if got != want {
		t.Errorf("%sgot %v, want %v", p, got, want)
	}
}
