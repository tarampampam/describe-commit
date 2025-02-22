package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gh.tarampamp.am/describe-commit/internal/config"
)

func TestDefaultDirPath(t *testing.T) {
	t.Parallel()

	if config.DefaultDirPath() == "" {
		t.Error("DefaultDirPath is empty")
	}
}

func TestFindIn(t *testing.T) {
	t.Parallel()

	var join = filepath.Join

	const (
		configFileName       = config.FileName
		hiddenConfigFileName = "." + configFileName
	)

	for name, tc := range map[string]struct {
		giveDirs  []string
		giveFiles []string
		giveDir   string
		want      []string
	}{
		"nothing found": {
			giveDirs: []string{
				"dir1", // /tmp/dir1
				"dir2", // /tmp/dir2
			},
			giveFiles: []string{
				"file1",               // /tmp/file1
				join("dir2", "file2"), // /tmp/dir2/file2
			},
			want: nil,
		},
		"root file only": {
			giveFiles: []string{
				"file1",              // /tmp/file1
				configFileName,       // /tmp/describe-commit.yml
				hiddenConfigFileName, // /tmp/.describe-commit.yml
			},
			want: []string{
				configFileName,
				hiddenConfigFileName,
			},
		},
		"to the root": {
			giveDirs: []string{
				"dir1/dir2", // /tmp/dir1/dir2
				"dir3",      // /tmp/dir3
			},
			giveFiles: []string{
				join("dir1", "file1"),                      // /tmp/dir1/file1
				join("dir1", "dir2", hiddenConfigFileName), // /tmp/dir1/dir2/.describe-commit.yml
				join("dir3", configFileName),               // /tmp/dir3/describe-commit.yml
				configFileName,                             // /tmp/describe-commit.yml
			},
			giveDir: join("dir1", "dir2"),
			want: []string{
				join("dir1", "dir2", hiddenConfigFileName),
				configFileName,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var tmpDir = t.TempDir()

			for _, giveDir := range tc.giveDirs {
				if err := os.MkdirAll(join(tmpDir, giveDir), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			for _, giveFile := range tc.giveFiles {
				if f, err := os.OpenFile(join(tmpDir, giveFile), os.O_CREATE, 0o644); err != nil {
					t.Fatal(err)
				} else {
					_ = f.Close()
				}
			}

			var found = config.FindIn(join(tmpDir, tc.giveDir))

			if got, want := len(found), len(tc.want); got != want {
				t.Errorf("unexpected number of found files (want %d): %d (%v)", got, want, found)
			}

			if len(tc.want) > 0 {
				for i, filePath := range found {
					var got, want = filePath, join(tmpDir, tc.want[i])

					if got != want {
						t.Errorf("unexpected file path (want %s): %s", want, got)
					}
				}
			}
		})
	}

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		if found := config.FindIn(""); found != nil {
			t.Errorf("unexpected found files: %v", found)
		}
	})

	t.Run("malformed directory", func(t *testing.T) {
		t.Parallel()

		if found := config.FindIn("\\*:**malformed"); found != nil {
			t.Errorf("unexpected found files: %v", found)
		}
	})
}
