package gitutil

import "testing"

func TestHasGitAndInit(t *testing.T) {
	dir := t.TempDir()
	if HasGit(dir) {
		t.Fatalf("expected no git before init")
	}
	if err := Init(dir); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if !HasGit(dir) {
		t.Fatalf("expected git after init")
	}
	t.Log("OK")
}
