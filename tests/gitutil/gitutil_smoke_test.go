package gitutil

import (
	"wip/internal/gitutil"

	"testing"
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
	t.Log("OK")
}
