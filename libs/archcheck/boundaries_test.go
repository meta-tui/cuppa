package archcheck

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// libsRoot is the libs/ directory, relative to this package.
const libsRoot = ".."

type importRule struct {
	name      string
	forbidden string
	// allowedIn lists the capabilities (first path segment under libs/) that may import it.
	allowedIn []string
}

// The headless-engine boundary: libs describe, edit, render and store designs
// and never depend on a front end. Only render may style text (Lip Gloss).
var rules = []importRule{
	{name: "libs stay headless: no Bubble Tea", forbidden: "charm.land/bubbletea", allowedIn: nil},
	{name: "libs stay headless: no Bubbles", forbidden: "charm.land/bubbles", allowedIn: nil},
	{name: "libs never import apps", forbidden: "github.com/meta-ui-labs/cuppa/apps/", allowedIn: nil},
	{name: "only render styles output", forbidden: "charm.land/lipgloss", allowedIn: []string{"render"}},
}

func TestLibsRespectTheHeadlessBoundary(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir(libsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, _ := filepath.Rel(libsRoot, path)
		capability := strings.Split(filepath.ToSlash(rel), "/")[0]
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			for _, r := range rules {
				if strings.HasPrefix(p, r.forbidden) && !contains(r.allowedIn, capability) {
					t.Errorf("%s: %s imports %s", r.name, filepath.ToSlash(rel), p)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
