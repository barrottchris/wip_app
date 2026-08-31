package frontend_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendRefactorEntryPoints(t *testing.T) {
	legacyFile := filepath.Join("..", "..", "frontend", "src", "main.js")
	if _, err := os.Stat(legacyFile); err == nil {
		t.Fatal("legacy frontend/src/main.js must be removed; the refactored module entry is frontend/src/js/main.js")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("checking legacy entry: %v", err)
	}

	moduleFile := filepath.Join("..", "..", "frontend", "src", "js", "main.js")
	content, err := os.ReadFile(moduleFile)
	if err != nil {
		t.Fatalf("missing refactored entry point: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "setPages") {
		t.Fatal("frontend/src/js/main.js must initialize the page map via setPages()")
	}
	if !strings.Contains(text, "\"app-git\": renderAppGitPage") {
		t.Fatal("frontend/src/js/main.js must register the Git page")
	}
	if !strings.Contains(text, "navigateTo(\"registry\")") {
		t.Fatal("frontend/src/js/main.js must start on the registry page")
	}
}
