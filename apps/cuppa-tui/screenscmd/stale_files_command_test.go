package screenscmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// The files of a page the design no longer has are listed when they are
// removed, and --keep-stale leaves them.
func TestTheFilesOfAGonePageAreListedAndCanBeKept(t *testing.T) {
	dir := t.TempDir()
	doc := design.NewDocument("App", 60, 12)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{X: 1, Y: 1, W: 20, H: 4}})
	doc.Pages = []design.Page{{Name: "Home"}, {Name: "Settings"}}
	doc.Export = design.ExportSettings{Layout: design.ExportSlices, Dir: "app", Module: "example.com/app"}
	file, err := disk.Save(filepath.Join(dir, "app"), doc)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	left := filepath.Join(dir, "app", "pages", "settings", "settings.go")
	if _, err := os.Stat(left); err != nil {
		t.Fatalf("the page was not written: %v", err)
	}

	doc.Pages = doc.Pages[:1]
	if _, err := disk.Save(filepath.Join(dir, "app"), doc); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := Run([]string{file, "--keep-stale"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "kept pages/settings/settings.go") {
		t.Errorf("the kept file is listed: %s", stdout.String())
	}
	if _, err := os.Stat(left); err != nil {
		t.Errorf("--keep-stale removed the file: %v", err)
	}

	stdout.Reset()
	if code := Run([]string{file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "removed pages/settings/settings.go") {
		t.Errorf("the removed file is listed: %s", stdout.String())
	}
	if _, err := os.Stat(left); err == nil {
		t.Error("the file of the page that is gone is still there")
	}
}
