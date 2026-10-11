package designcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
)

func buildDesign(t *testing.T, dir, name, w, h, text, sampleValue string) string {
	t.Helper()
	json := `{"document": {"name": "` + name + `", "width": ` + w + `, "height": ` + h + `,
	  "samples": {"Rows": "` + sampleValue + `"},
	  "keys": [{"key": "esc", "event": "Back` + name + `", "label": "back"}],
	  "nodes": [{"id": "n1", "component": "lipgloss.label", "name": "Title", "rect": {"X": 1, "Y": 1, "W": 20, "H": 1},
	   "props": {"text": "` + text + `"}, "layout": {"h": "$Rows"}}]}}`
	src := filepath.Join(dir, name+".json")
	if err := os.WriteFile(src, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, name+".cuppa")
	if code, o, e := run(t, "build", src, out); code != 0 {
		t.Fatalf("build %s: %d\n%s%s", name, code, o, e)
	}
	return out
}

// Merging designs gives one page each, in the order given, and a page keeps its
// own samples and keys.
func TestMergingDesignsKeepsEachPagesSamplesAndKeys(t *testing.T) {
	dir := t.TempDir()
	a := buildDesign(t, dir, "Setup", "100", "30", "setup", "3")
	b := buildDesign(t, dir, "Run", "120", "36", "run", "9")
	out := filepath.Join(dir, "all.cuppa")
	code, o, e := run(t, "merge", a, b, "-o", out, "--name", "App", "--folder", "Setup=setup/prefs")
	if code != 0 {
		t.Fatalf("merge: %d\n%s%s", code, o, e)
	}
	doc, err := disk.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name != "App" || doc.PageCount() != 2 || doc.Width != 120 || doc.Height != 36 {
		t.Fatalf("got %q with %d pages, %dx%d", doc.Name, doc.PageCount(), doc.Width, doc.Height)
	}
	screens := doc.Screens()
	if screens[0].Name != "Setup" || screens[0].Folder != "setup/prefs" || screens[0].Samples["Rows"] != "3" {
		t.Errorf("first page: %+v", screens[0])
	}
	if screens[1].Name != "Run" || screens[1].Samples["Rows"] != "9" || len(screens[1].Keys) != 1 || screens[1].Keys[0].Event != "BackRun" {
		t.Errorf("second page: name %q samples %v keys %v", screens[1].Name, screens[1].Samples, screens[1].Keys)
	}
	if screens[0].Keys[0].Event != "BackSetup" {
		t.Errorf("first page keys: %v", screens[0].Keys)
	}
}

func TestMergingRefusesTwoDesignsWithOneName(t *testing.T) {
	dir := t.TempDir()
	a := buildDesign(t, dir, "Same", "100", "30", "a", "1")
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	b := buildDesign(t, sub, "Same", "100", "30", "b", "2")
	code, _, e := run(t, "merge", a, b, "-o", filepath.Join(dir, "x.cuppa"))
	if code == 0 || !strings.Contains(e, "Same") {
		t.Fatalf("merge: %d %s", code, e)
	}
}
