package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Log returns the commit log of the repository limited to the specified number of commits.
func Log(ctx context.Context, dirPath string, len int) (string, error) {
	// ensure git is installed and available to run
	gitFilePath, lookErr := binPath()
	if lookErr != nil {
		return "", lookErr
	}

	// get the log
	var cmd = exec.CommandContext(ctx, gitFilePath, "log",
		"--format=%s",
		fmt.Sprintf("--max-count=%d", len),
		"--no-color",
	)

	cmd.Dir = dirPath

	var stdOut, stdErr bytes.Buffer

	stdOut.Grow(1024 * 2) //nolint:mnd // 2KB

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		if stdErr.Len() > 0 {
			err = fmt.Errorf("%s: %w", stdErrToString(stdErr.String()), err)
		}

		return "", fmt.Errorf("git log failed: %w", err)
	}

	return stdOut.String(), nil
}
