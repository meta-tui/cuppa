package gosource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// The files a page that is gone leaves are reported before they are removed,
// and a write that keeps them leaves them (and only them) in place.
func TestTheFilesOfAGonePageAreReportedAndCanBeKept(t *testing.T) {
	cat := standard.Default()
	s := design.ExportSettings{Layout: design.ExportSlices, Module: "example.com/shop"}
	dir := t.TempDir()
	first, err := GenerateExport(pagedApp(), cat, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteExport(dir, s, first); err != nil {
		t.Fatal(err)
	}
	if got := ObsoleteExport(dir, s, first); len(got) != 0 {
		t.Fatalf("a fresh export has nothing to remove: %v", got)
	}

	doc := pagedApp()
	doc.Pages = doc.Pages[:1]
	doc.Page = 0
	second, err := GenerateExport(doc, cat, s, nil)
	if err != nil {
		t.Fatal(err)
	}
	gone := ObsoleteExport(dir, s, second)
	if len(gone) == 0 || !strings.Contains(strings.Join(gone, " "), "settings") {
		t.Fatalf("the page that is gone is reported: %v", gone)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "settings", "settings.cuppa.go")); err != nil {
		t.Fatalf("reporting must not remove anything: %v", err)
	}

	if err := WriteExportKeeping(dir, s, second, gone); err != nil {
		t.Fatal(err)
	}
	for _, name := range gone {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			t.Errorf("%s was kept and is gone: %v", name, err)
		}
	}
	// Exporting again without keeping removes them.
	if err := WriteExport(dir, s, second); err != nil {
		t.Fatal(err)
	}
	for _, name := range gone {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err == nil {
			t.Errorf("%s should have been removed", name)
		}
	}
}

// The one-folder layout writes files with no header: it cannot tell which are
// its own, so it reports nothing and removes nothing.
func TestTheOneFolderLayoutHasNothingToRemove(t *testing.T) {
	s := design.ExportSettings{Layout: design.ExportFlat}
	if got := ObsoleteExport(t.TempDir(), s, Project{}); got != nil {
		t.Fatalf("got %v", got)
	}
}
