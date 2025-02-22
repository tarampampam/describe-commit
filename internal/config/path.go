package config

import (
	"os"
	"path/filepath"
)

// FileName holds the name of the configuration file.
const FileName = "describe-commit.yml"

// DefaultDirPathEnvName used to override the default directory path (useful for docs generation purposes).
const DefaultDirPathEnvName = "DEFAULT_CONFIG_FILE_DIR"

// DefaultDirPath returns the default directory path where the configuration file is looked for by default.
// Only in case of exception, this function returns an empty string.
func DefaultDirPath() string {
	if v, ok := os.LookupEnv(DefaultDirPathEnvName); ok {
		return v
	}

	if v := osSpecificConfigDirPath(); v != "" {
		return v
	}

	if v, err := os.Getwd(); err == nil {
		return v // fallback to the current working directory
	}

	return "" // no default path
}

// FindIn searches for the configuration file in the specified directory and its parent directories up to the root.
// The search order is maintained, starting from the specified directory and moving toward the root.
// The returned slice contains the absolute paths of the found files.
//
// For example, if for the following directory structure:
//
//	.
//	├── dir1
//	│   ├── dir2
//	│   │   ├── dir3
//	│   │   │   └── describe-commit.yml
//	│   │   └── dir4
//	│   │       └── describe-commit.yml
//	│   ├── .describe-commit.yml
//	│   └── describe-commit.yml
//	└── describe-commit.yml
//
// When this function is called with the path of `/dir1/dir2/dir3`, it will return the following slice:
//
//	[
//		"/dir1/dir2/dir3/describe-commit.yml",
//		"/dir1/describe-commit.yml", // non-hidden file has priority
//		"/dir1/.describe-commit.yml",
//		"/describe-commit.yml",
//	]
func FindIn(dirPath string) []string {
	if dirPath == "" {
		return nil // no directory provided
	}

	// convert to absolute path
	if !filepath.IsAbs(dirPath) {
		if abs, err := filepath.Abs(dirPath); err == nil {
			dirPath = abs
		} else {
			return nil // can't convert to absolute path
		}
	}

	if stat, err := os.Stat(dirPath); err != nil || !stat.IsDir() {
		return nil // is not a directory, or doesn't exist
	}

	var searchIn = []string{dirPath} // include the provided directory

	var (
		current = filepath.Dir(dirPath)
		parent  string
	)

	for { // get directories from the provided directory up to the root
		parent = filepath.Dir(current)

		// stop when the parent is the same as the current (root reached)
		if parent == current {
			break
		}

		searchIn = append(searchIn, current)
		current = parent
	}

	var (
		fileNames = [...]string{FileName, "." + FileName} // allow the file to be hidden
		found     []string
	)

	for _, dir := range searchIn {
		for _, fileName := range fileNames {
			var path = filepath.Join(dir, fileName)

			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				found = append(found, path)
			}
		}
	}

	return found
}
