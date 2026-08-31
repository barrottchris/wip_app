package gitutil

import (
	"os"
	"os/exec"
	"testing"

	"wip/internal/gitutil"
)

func TestHasGitAndInit(t *testing.T) {
	dir := t.TempDir()
	if gitutil.HasGit(dir) {
		t.Fatalf("expected no git before init")
	}
	if err := gitutil.Init(dir); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if !gitutil.HasGit(dir) {
		t.Fatalf("expected git after init")
	}
	if branch, err := gitutil.DefaultBranch(dir); err != nil {
		t.Fatalf("default branch check failed: %v", err)
	} else if branch == "" {
		t.Fatal("expected a default branch name from the initialized repo")
	}
	t.Log("OK")
}

func TestDefaultBranchAfterCommit(t *testing.T) {
	dir := t.TempDir()
	if err := gitutil.Init(dir); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if err := os.WriteFile(dir+"/hello.txt", []byte("hello"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("config user.name: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("config user.email: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "add", "hello.txt").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "commit", "-m", "init").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	branch, err := gitutil.DefaultBranch(dir)
	if err != nil {
		t.Fatalf("default branch detection failed: %v", err)
	}
	if branch != "master" && branch != "main" {
		t.Fatalf("unexpected default branch %q", branch)
	}
}
