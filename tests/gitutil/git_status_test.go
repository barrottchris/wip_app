package gitutil

import (
	"testing"

	"wip/internal/gitutil"
)

func TestRepositoryName(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		path      string
		want      string
	}{
		{name: "https remote", remoteURL: "https://github.com/example/demo.git", path: `C:\\Work\\fallback`, want: "demo"},
		{name: "ssh remote", remoteURL: "git@github.com:example/demo.git", path: `C:\\Work\\fallback`, want: "demo"},
		{name: "folder fallback", path: `C:\\Work\\fallback`, want: "fallback"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := gitutil.RepositoryName(test.remoteURL, test.path); got != test.want {
				t.Fatalf("RepositoryName(%q, %q) = %q; want %q", test.remoteURL, test.path, got, test.want)
			}
		})
	}
}
