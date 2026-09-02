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
	if !strings.Contains(text, "navigateTo(\"home\")") {
		t.Fatal("frontend/src/js/main.js must start on the home page")
	}

	indexFile := filepath.Join("..", "..", "frontend", "src", "index.html")
	indexContent, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("missing frontend shell: %v", err)
	}
	if !strings.Contains(string(indexContent), `id="brand-home"`) {
		t.Fatal("frontend brand must provide a home navigation target")
	}
}

func TestArchiveWarningUsesInAppPrompt(t *testing.T) {
	appDetailFile := filepath.Join("..", "..", "frontend", "src", "js", "pages", "appDetail.js")
	content, err := os.ReadFile(appDetailFile)
	if err != nil {
		t.Fatalf("missing app detail page: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "confirm(") {
		t.Fatal("archive flow must not use browser confirm() dialogs; use an in-app warning panel")
	}
	if !strings.Contains(text, "archive-confirm") {
		t.Fatal("archive flow must render an in-app confirmation panel")
	}
}
