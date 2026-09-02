// Package gitutil wraps the local `git` command for the operations
// onboarding needs. MVP scope only: detecting whether a folder is already
// a repo, and initializing one if the user opts in. Reading real branch/
// commit info (replacing the placeholder data in the app store) is a
// separate, later piece of work.
package gitutil

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

// RepositoryName derives a short repository name from a remote URL, falling
// back to the local folder name when the repo has no usable remote.
func RepositoryName(remoteURL, path string) string {
	remoteURL = strings.TrimSuffix(strings.TrimSpace(remoteURL), "/")
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	if remoteURL != "" {
		if parsed, err := url.Parse(remoteURL); err == nil && parsed.Path != "" {
			if name := filepath.Base(parsed.Path); name != "." && name != string(filepath.Separator) && name != "" {
				return name
			}
		}
		if separator := strings.LastIndexAny(remoteURL, "/:"); separator >= 0 && separator+1 < len(remoteURL) {
			return remoteURL[separator+1:]
		}
	}
	if path != "" {
		return filepath.Base(filepath.Clean(path))
	}
	return ""
}

// LastCommitAt reads the committer timestamp of the latest commit on the
// currently checked-out branch.
func LastCommitAt(path string) (time.Time, error) {
	if !HasGit(path) {
		return time.Time{}, fmt.Errorf("not a git repo: %s", path)
	}
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%cI", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
}

// Init runs `git init` in path. Only called after the user explicitly
// confirms — WIP never initializes git silently (see mvp-scope.md).
func Init(path string) error {
	cmd := exec.Command("git", "init", path)
	return cmd.Run()
}
