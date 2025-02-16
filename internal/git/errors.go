package git

import "strings"

// stdErrToString converts the standard error output into a more readable format (mot much, but anyway).
func stdErrToString(stdErr string) string {
	if len(stdErr) == 0 {
		return ""
	}

	var lines = strings.Split(stdErr, "\n")

	// remove empty lines
	for i := 0; i < len(lines); i++ {
		if len(strings.TrimSpace(lines[i])) == 0 {
			lines = append(lines[:i], lines[i+1:]...)
			i--
		}
	}

	return strings.Join(lines, "; ")
}
