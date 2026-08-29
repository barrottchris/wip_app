// Package gitutil wraps the local `git` command for the operations
// onboarding needs. MVP scope only: detecting whether a folder is already
// a repo, and initializing one if the user opts in. Reading real branch/
// commit info (replacing the placeholder data in the app store) is a
// separate, later piece of work.
package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
)

// HasGit reports whether path already contains a .git directory.
func HasGit(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

// Init runs `git init` in path. Only called after the user explicitly
// confirms — WIP never initializes git silently (see mvp-scope.md).
func Init(path string) error {
	cmd := exec.Command("git", "init", path)
	return cmd.Run()
}
