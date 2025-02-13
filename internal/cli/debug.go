package cli

import (
	"fmt"
	"os"
	"strconv"
)

const debugEnvName = "DEBUG" // environment variable name to enable debug output

//nolint:gochecknoglobals,nlreturn
var isDebuggingOn = func() (b bool) { b, _ = strconv.ParseBool(os.Getenv(debugEnvName)); return }()

// debugf is a helper function to print debug information to the stderr.
func debugf(format string, args ...any) {
	if isDebuggingOn {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("# [trace] %s\n", format), args...)
	}
}
