// Package gitutil wraps the local `git` command for the operations
// onboarding needs. MVP scope only: detecting whether a folder is already
// a repo, and initializing one if the user opts in. Reading real branch/
// commit info (replacing the placeholder data in the app store) is a
// separate, later piece of work.
package gitutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HasGit reports whether path already contains a .git directory.
func HasGit(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

// DefaultBranch reads the repository's current default branch name, such as
// "main" or "master". This comes from the repo itself, not from seeded app data.
func DefaultBranch(path string) (string, error) {
	if !HasGit(path) {
		return "", fmt.Errorf("not a git repo: %s", path)
	}

	cmd := exec.Command("git", "-C", path, "symbolic-ref", "--quiet", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
		out, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(out)), nil
}

// RemoteURL returns the configured origin URL for the repo, if any.
func RemoteURL(path string) (string, error) {
	if !HasGit(path) {
		return "", fmt.Errorf("not a git repo: %s", path)
	}
	cmd := exec.Command("git", "-C", path, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Init runs `git init` in path. Only called after the user explicitly
// confirms — WIP never initializes git silently (see mvp-scope.md).
func Init(path string) error {
	cmd := exec.Command("git", "init", path)
	return cmd.Run()
}
