package cmd_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// assertNoError fails the test if err is not nil, indicating an unexpected error occurred.
func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// assertErrorContains checks if the given error message contains the expected substring.
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Errorf("expected error, got nil")

		return
	}

	assertContains(t, err.Error(), substr)
}

// assertContains checks if the given string contains at least one of the expected substrings.
func assertContains(t *testing.T, got string, want ...string) {
	t.Helper()

	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("expected %q to contain %q", got, w)
		}
	}
}

// assertSlicesEqual checks if two slices are equal in both length and content.
// It provides an optional message prefix for more detailed failure messages.
func assertSlicesEqual[T comparable](t *testing.T, got, want []T, msgPrefix ...string) {
	t.Helper()

	var p string

	if len(msgPrefix) > 0 {
		p = msgPrefix[0] + ": "
	}

	if len(got) != len(want) {
		t.Errorf("%slices have different lengths: got %d, want %d", p, len(got), len(want))

		return
	}

	for i := range got {
		assertEqual(t, got[i], want[i], fmt.Sprintf("%sindex %d", p, i))
	}
}

// assertEqual checks if two values of a comparable type are equal.
// If they are not, the test fails and includes an optional message prefix.
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

// seededRand is a global pseudo-random number generator seeded with the current time.
// It ensures different outputs on each program run.
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// randomString generates a random alphanumeric string of the specified length.
// Uses a predefined character set and a seeded random source for variability.
func randomString(strLen int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var b = make([]byte, strLen)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
