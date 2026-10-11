package fileflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// exportedWithTwoPages exports a design of two pages as slices, then removes
// the second page and returns the file that page left.
func exportedWithTwoPages(t *testing.T) (f *Flow, settings string) {
	t.Helper()
	f, ed, dir := setup(t)
	if err := ed.AddPage("Settings"); err != nil {
		t.Fatal(err)
	}
	app := filepath.Join(dir, "shop")
	if err := ed.SetExportSettings(design.ExportSettings{Layout: design.ExportSlices, Dir: app, Module: "example.com/shop"}); err != nil {
		t.Fatal(err)
	}
	f.ExportGoSource()
	press(f, true, false) // Export
	if f.Modal() == nil || !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "Written to") {
		t.Fatalf("the first export writes: %v", f.Modal().Lines())
	}
	press(f, true, false)
	left := filepath.Join(app, "pages", "settings", "settings.go")
	if _, err := os.Stat(left); err != nil {
		t.Fatalf("the page was not written: %v", err)
	}
	if err := ed.DeletePage(1); err != nil {
		t.Fatal(err)
	}
	return f, left
}

func TestExportAsksBeforeDeletingTheFilesOfARemovedPage(t *testing.T) {
	f, left := exportedWithTwoPages(t)
	f.ExportGoSource()
	press(f, true, false) // Export
	text := stripped(strings.Join(f.Modal().Lines(), "\n"))
	if !strings.Contains(text, "Delete Go files") || !strings.Contains(text, "settings.go") {
		t.Fatalf("the files to delete are listed: %s", text)
	}
	if _, err := os.Stat(left); err != nil {
		t.Fatal("nothing is deleted before the answer")
	}
	press(f, true, false) // Delete, the first button
	if _, err := os.Stat(left); err == nil {
		t.Error("Delete left the file")
	}
	if !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "Written to") {
		t.Error("the export goes on after Delete")
	}
}

func TestKeepLeavesTheFilesAndStillExports(t *testing.T) {
	f, left := exportedWithTwoPages(t)
	f.ExportGoSource()
	press(f, true, false) // Export
	clickButton(t, f, "Keep")
	if _, err := os.Stat(left); err != nil {
		t.Errorf("Keep removed the file: %v", err)
	}
	if !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "were kept") {
		t.Errorf("the notice says the files were kept: %v", f.Modal().Lines())
	}
}

func TestCancelWritesNothingAndDeletesNothing(t *testing.T) {
	f, left := exportedWithTwoPages(t)
	f.ExportGoSource()
	press(f, true, false) // Export
	clickButton(t, f, "Cancel")
	if _, err := os.Stat(left); err != nil {
		t.Errorf("Cancel removed the file: %v", err)
	}
	if f.Modal() != nil {
		t.Errorf("Cancel leaves a dialog: %v", f.Modal().Lines())
	}
}
