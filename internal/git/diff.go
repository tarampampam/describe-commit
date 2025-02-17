package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Diff returns the diff of the staged changes or changes between the index and the working tree.
func Diff(ctx context.Context, dirPath string) (string, error) {
	// ensure git is installed and available to run
	gitFilePath, lookErr := binPath()
	if lookErr != nil {
		return "", lookErr
	}

	// get the diff
	var cmd = exec.CommandContext(ctx, gitFilePath, "diff",
		"--cached",                 // show all staged changes or changes between the index and the working tree
		"--ignore-submodules=all",  // ignore changes to submodules
		"--diff-algorithm=minimal", // use the minimal diff algorithm
		"--no-ext-diff",            // do not use external diff helper
		"--ignore-all-space",       // ignore whitespace when comparing lines
		"--ignore-blank-lines",     // ignore changes whose lines are all blank
		"--no-color",               // do not use any color in the output
		"--patch",                  // generate patch (unified diff) format
		"--",
		":(exclude)*.sum",  // exclude .sum files
		":(exclude)*.lock", // exclude .lock files
		":(exclude)*.log",  // exclude .log files
		":(exclude)*.out",  // exclude .out files
		":(exclude)*.tmp",  // exclude .tmp files
		":(exclude)*.bak",  // exclude .bak files
		":(exclude)*.swp",  // exclude .swp files
		":(exclude)*.env",  // exclude .env files
	)

	cmd.Dir = dirPath
	cmd.Env = []string{
		"LC_ALL=C", "LANG=C", // forces the system to use the "C" (POSIX) locale, English-based output with no localization
		"NO_COLOR=1",            // disables colored output
		"GIT_CONFIG_NOSYSTEM=1", // do not use the system-wide configuration file
	}

	var stdOut, stdErr bytes.Buffer

	stdOut.Grow(1024 * 8) //nolint:mnd // 8KB

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		if stdErr.Len() > 0 {
			err = fmt.Errorf("%s: %w", stdErrToString(stdErr.String()), err)
		}

		return "", fmt.Errorf("git diff failed: %w", err)
	}

	return stdOut.String(), nil
}
