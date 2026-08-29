package fsbrowse

import (
	"runtime"
	"testing"
)

func TestListDefaultsToRoot(t *testing.T) {
	listing, err := List("")
	if err != nil {
		t.Fatalf("List(\"\") failed: %v", err)
	}
	want := "/"
	if runtime.GOOS == "windows" {
		want = `C:\`
	}
	if listing.CurrentPath != want {
		t.Errorf("expected default path %q, got %q", want, listing.CurrentPath)
	}
	t.Logf("OK: defaulted to %s", listing.CurrentPath)
}
