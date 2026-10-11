package gosource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// A page is written as <page>.cuppa.go. The <page>.go an earlier export wrote
// goes (it has the header); a <page>.go the person wrote stays, untouched.
func TestGeneratedFilesHaveTheCuppaSuffixAndTheOldOnesAreMigrated(t *testing.T) {
	s := design.ExportSettings{Layout: design.ExportSlices, Module: "example.com/shop"}
	p, err := GenerateExport(pagedApp(), standard.Default(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"pages/home/home.cuppa.go", "pages/settings/settings.cuppa.go"} {
		src, ok := p.Files[name]
		if !ok {
			t.Fatalf("no %s in %v", name, keys(p.Files))
		}
		if !strings.HasPrefix(src, GeneratedHeader) {
			t.Errorf("%s lacks the generated header", name)
		}
	}

	dir := t.TempDir()
	write := func(rel, text string) {
		t.Helper()
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("pages/home/home.go", GeneratedHeader+"\npackage home\n") // an earlier export
	write("pages/settings/settings.go", "package settings\n\n// mine\n")
	if err := WriteExport(dir, s, p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "home", "home.go")); err == nil {
		t.Error("the page the earlier export wrote is still there")
	}
	if data, err := os.ReadFile(filepath.Join(dir, "pages", "settings", "settings.go")); err != nil || !strings.Contains(string(data), "// mine") {
		t.Errorf("a hand-written file was touched: %v", err)
	}
	for _, name := range []string{"pages/home/home.cuppa.go", "pages/settings/settings.cuppa.go"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			t.Errorf("%s was not written", name)
		}
	}
}
