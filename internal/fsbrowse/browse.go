// Package fsbrowse lets the frontend browse the local filesystem via the Go
// backend — necessary because a browser's file picker deliberately can't
// return a real, absolute OS path for security reasons. Since our backend
// already runs locally with full disk access, it lists directories itself
// and the frontend just renders whatever comes back.
package fsbrowse

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

// Entry is a single browsable directory.
type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Listing is what the folder-picker UI renders for one screen: the current
// path, its parent (for "up" navigation, empty if there isn't one), and the
// subdirectories inside it.
type Listing struct {
	CurrentPath string  `json:"currentPath"`
	ParentPath  string  `json:"parentPath"`
	Directories []Entry `json:"directories"`
}

// List returns the subdirectories of path. Files are intentionally excluded
// — this picker is for choosing a project folder, not a file.
func List(path string) (Listing, error) {
	if path == "" {
		// Default to a drive/filesystem root for obvious, unbiased
		// navigation — not the current user's home folder, which buries
		// the picker several clicks away from anywhere useful (e.g. C:\Dev).
		if runtime.GOOS == "windows" {
			path = `C:\`
		} else {
			path = "/"
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return Listing{}, err
	}

	var dirs []Entry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Skip hidden/system-ish folders to keep the picker uncluttered.
		if len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}
		dirs = append(dirs, Entry{
			Name: e.Name(),
			Path: filepath.Join(path, e.Name()),
		})
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })

	parent := filepath.Dir(path)
	if parent == path {
		parent = "" // already at a filesystem root
	}

	return Listing{
		CurrentPath: path,
		ParentPath:  parent,
		Directories: dirs,
	}, nil
}
