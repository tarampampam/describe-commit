//go:build generate

package main

import (
	"os"

	cliDocs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"

	ourApp "gh.tarampamp.am/describe-commit/internal/cli"
)

func main() {
	const readmePath = "../../README.md"

	if stat, err := os.Stat(readmePath); err == nil && stat.Mode().IsRegular() {
		var app = ourApp.NewApp()

		// patch the default value of the "config-file" flag to avoid OS-dependent paths
		for _, flag := range app.Flags {
			if f, ok := flag.(*cli.StringFlag); ok && f.Name == "config-file" {
				f.Value = "/depends/on/your-os/describe-commit.yml"
				break
			}
		}

		if err = cliDocs.ToTabularToFileBetweenTags(app, "describe-commit", readmePath); err != nil {
			panic(err)
		} else {
			println("✔ cli docs updated successfully")
		}
	} else if err != nil {
		println("⚠ readme file not found, cli docs not updated:", err.Error())
	}
}
