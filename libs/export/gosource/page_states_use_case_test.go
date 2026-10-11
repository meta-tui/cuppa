package gosource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestAPageIsNewThenCurrentThenChanged(t *testing.T) {
	s := design.ExportSettings{Layout: design.ExportSlices, Module: "example.com/shop"}
	p, err := GenerateExport(pagedApp(), standard.Default(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for page, got := range PageStates(dir, p) {
		if got != PageNew {
			t.Errorf("%s before the export: %s", page, got)
		}
	}
	if err := WriteExport(dir, s, p); err != nil {
		t.Fatal(err)
	}
	for page, got := range PageStates(dir, p) {
		if got != PageCurrent {
			t.Errorf("%s after the export: %s", page, got)
		}
	}
	doc := pagedApp()
	doc.Pages[1].Nodes[0].Name = "Renamed"
	q, err := GenerateExport(doc, standard.Default(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := PageStates(dir, q)
	if got["Settings"] != PageChanged || got["Home"] != PageCurrent {
		t.Fatalf("after editing Settings: %v", got)
	}
	// A page's file that is gone is a change too.
	if err := os.Remove(filepath.Join(dir, "pages", "home", "home.cuppa.go")); err != nil {
		t.Fatal(err)
	}
	if got := PageStates(dir, q); got["Home"] != PageNew {
		t.Fatalf("Home without its file: %v", got)
	}
}
