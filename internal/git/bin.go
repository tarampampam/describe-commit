package git

import (
	"fmt"
	"os/exec"
	"runtime"
)

// binPath returns the path to the git binary.
func binPath() (string, error) {
	var name = "git"

	if runtime.GOOS == "windows" {
		name = "git.exe"
	}

	p, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found: %w", name, err)
	}

	return p, nil
}
