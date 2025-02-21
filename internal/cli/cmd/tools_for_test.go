package cmd_test

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Errorf("expected error, got nil")

		return
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("expected error to contain %q, got %v", substr, err)
	}
}

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

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomString(strLen int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var b = make([]byte, strLen)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
