package fsbrowse

import (
	"wip/internal/fsbrowse"

	"os"
	"testing"
)

func TestListSmoke(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir+"/subdir1", 0o755)
	os.MkdirAll(dir+"/subdir2", 0o755)
	os.MkdirAll(dir+"/.hidden", 0o755)

	listing, err := fsbrowse.List(dir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(listing.Directories) != 2 {
		t.Fatalf("expected 2 visible dirs (hidden excluded), got %d: %v", len(listing.Directories), listing.Directories)
	}
	if listing.CurrentPath != dir {
		t.Fatalf("expected currentPath %s, got %s", dir, listing.CurrentPath)
	}
	t.Logf("OK: %+v", listing)
}
