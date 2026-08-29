package app

import (
	"testing"

	"wip/internal/app"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"My Cool App":  "my-cool-app",
		"SRE Agent!!":  "sre-agent",
		"  spaced  ":   "spaced",
		"Already-fine": "already-fine",
	}
	for input, want := range cases {
		got := app.Slugify(input)
		if got != want {
			t.Errorf("Slugify(%q) = %q, want %q", input, got, want)
		}
	}
}
