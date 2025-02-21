package debug

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
)

var Enabled atomic.Bool //nolint:gochecknoglobals

// Printf is a helper function to print debug information to the stderr.
func Printf(format string, args ...any) {
	if Enabled.Load() {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("# [debug] %s\n", format), args...)
	}
}

func init() { //nolint:gochecknoinits
	const debugEnvName = "DEBUG" // environment variable name to enable debug output

	if v, ok := os.LookupEnv(debugEnvName); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			Enabled.Store(b)
		}
	}
}
