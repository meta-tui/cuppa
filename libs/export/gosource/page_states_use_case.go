package gosource

import (
	"os"
	"path/filepath"
	"strings"
)

// Page states: how a page's exported files in a folder compare with what the
// design would export now.
const (
	// PageCurrent: the page's files are what the design exports.
	PageCurrent = "current"
	// PageChanged: a file of the page differs, or part of it is missing.
	PageChanged = "changed"
	// PageNew: none of the page's files is in the folder.
	PageNew = "new"
)

// PageStates says, for each page of the project, whether its files in dir are
// current, changed since the export or not exported yet. Line endings are not
// compared. Nothing is written.
func PageStates(dir string, p Project) map[string]string {
	same := func(a, b string) bool {
		return strings.ReplaceAll(a, "\r\n", "\n") == strings.ReplaceAll(b, "\r\n", "\n")
	}
	out := make(map[string]string, len(p.PageFiles))
	for page, files := range p.PageFiles {
		found, differs := 0, false
		for _, name := range files {
			data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
			if err != nil {
				differs = true
				continue
			}
			found++
			if !same(string(data), p.Files[name]) {
				differs = true
			}
		}
		switch {
		case found == 0:
			out[page] = PageNew
		case differs:
			out[page] = PageChanged
		default:
			out[page] = PageCurrent
		}
	}
	return out
}
