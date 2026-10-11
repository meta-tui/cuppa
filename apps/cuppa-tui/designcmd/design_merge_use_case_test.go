package designcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
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

// --master lifts what is the same on enough pages; a page without it hides it,
// and every screen still has the components it had.
func TestMergingWithAMasterLiftsWhatPagesShare(t *testing.T) {
	dir := t.TempDir()
	var files []string
	for _, name := range []string{"One", "Two", "Three", "Four"} {
		files = append(files, buildDesign(t, dir, name, "100", "30", name+" text", "1"))
	}
	plain := filepath.Join(dir, "plain.cuppa")
	if code, o, e := run(t, append(append([]string{"merge"}, files...), "-o", plain)...); code != 0 {
		t.Fatalf("merge: %d\n%s%s", code, o, e)
	}
	// "Title" has a different text on every design, so it is not the same.
	out := filepath.Join(dir, "master.cuppa")
	code, o, e := run(t, append(append([]string{"merge"}, files...), "-o", out, "--master")...)
	if code != 0 {
		t.Fatalf("merge: %d\n%s%s", code, o, e)
	}
	if !strings.Contains(o, "nothing is on 3 pages") {
		t.Fatalf("different texts were lifted:\n%s", o)
	}
}

func TestTheMasterKeepsEveryScreenWhole(t *testing.T) {
	dir := t.TempDir()
	// Four designs whose Title is the same (a property bound to an input).
	var files []string
	for _, name := range []string{"One", "Two", "Three", "Four"} {
		json := `{"document": {"name": "` + name + `", "width": 100, "height": 30,
		  "nodes": [
		   {"id": "n1", "component": "lipgloss.label", "name": "Title", "rect": {"X": 0, "Y": 0, "W": 100, "H": 1},
		    "props": {"text": "` + name + `"}, "bind": {"text": "Title"}},
		   {"id": "n2", "component": "lipgloss.box", "name": "` + name + ` box", "rect": {"X": 2, "Y": 3, "W": 20, "H": 5}}]}}`
		src := filepath.Join(dir, name+".json")
		if err := os.WriteFile(src, []byte(json), 0o644); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, name+".cuppa")
		if code, o, e := run(t, "build", src, out); code != 0 {
			t.Fatalf("build: %d %s%s", code, o, e)
		}
		files = append(files, out)
	}
	out := filepath.Join(dir, "all.cuppa")
	code, o, e := run(t, append(append([]string{"merge"}, files...), "-o", out, "--master")...)
	if code != 0 || !strings.Contains(o, `lifted "Title"`) {
		t.Fatalf("merge: %d\n%s%s", code, o, e)
	}
	doc, err := disk.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if doc.MasterIndex() < 0 || len(doc.MasterNodes()) != 1 {
		t.Fatalf("master: index %d nodes %d", doc.MasterIndex(), len(doc.MasterNodes()))
	}
	screens := doc.Screens()
	if len(screens) != 4 {
		t.Fatalf("screens = %d", len(screens))
	}
	for _, s := range screens {
		if len(s.Nodes) != 2 {
			t.Fatalf("screen %s has %d nodes", s.Name, len(s.Nodes))
		}
		got := []string{s.Nodes[0].Name, s.Nodes[1].Name}
		if got[0] != "Title" || got[1] != s.Name+" box" {
			t.Errorf("screen %s: %v", s.Name, got)
		}
	}
}

func TestCheckingReportsANameThePageAndTheMasterShare(t *testing.T) {
	dir := t.TempDir()
	doc := design.NewDocument("App", 60, 10)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Title", Rect: design.Rect{W: 10, H: 1}})
	doc.Pages = []design.Page{
		{Name: "Home"},
		{Name: "Master", Master: true, Nodes: []design.Node{{ID: "n8", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{W: 10, H: 1}}}},
	}
	file, err := disk.Save(filepath.Join(dir, "app"), doc)
	if err != nil {
		t.Fatal(err)
	}
	code, o, _ := run(t, "check", file)
	if code != 1 || !strings.Contains(o, `"Title" on the page and on the master`) {
		t.Fatalf("check: %d\n%s", code, o)
	}
}

// A lifted component keeps what each page had for its bound text, as the
// page's own value, and a variant that is left says what differs.
func TestLiftingKeepsEachPagesBoundTextAndSaysWhatDiffers(t *testing.T) {
	dir := t.TempDir()
	var files []string
	for _, name := range []string{"One", "Two", "Three", "Four"} {
		show := ""
		if name == "Four" {
			show = `, "showIf": "Show keys"`
		}
		json := `{"document": {"name": "` + name + `", "width": 100, "height": 30,
		  "nodes": [
		   {"id": "n1", "component": "lipgloss.label", "name": "Title", "rect": {"X": 0, "Y": 0, "W": 100, "H": 1},
		    "props": {"text": "App · ` + name + `"}, "bind": {"text": "Title"}},
		   {"id": "n2", "component": "lipgloss.label", "name": "Footer", "rect": {"X": 0, "Y": 29, "W": 100, "H": 1},
		    "props": {"text": "keys"}` + show + `}]}}`
		src := filepath.Join(dir, name+".json")
		if err := os.WriteFile(src, []byte(json), 0o644); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, name+".cuppa")
		if code, o, e := run(t, "build", src, out); code != 0 {
			t.Fatalf("build: %d %s%s", code, o, e)
		}
		files = append(files, out)
	}
	out := filepath.Join(dir, "all.cuppa")
	code, o, e := run(t, append(append([]string{"merge"}, files...), "-o", out, "--master")...)
	if code != 0 {
		t.Fatalf("merge: %d\n%s%s", code, o, e)
	}
	if !strings.Contains(o, `lifted "Footer"`) || !strings.Contains(o, `lifted "Title"`) {
		t.Fatalf("report:\n%s", o)
	}
	doc, err := disk.Load(out)
	if err != nil {
		t.Fatal(err)
	}
	screens := doc.Screens()
	for i, name := range []string{"One", "Two", "Three"} {
		var title string
		for _, n := range screens[i].Nodes {
			if n.Name == "Title" {
				title = n.Props["text"]
			}
		}
		if title != "App · "+name {
			t.Errorf("page %s draws the title %q", name, title)
		}
	}
	// Four has a different Footer (it has a condition): it keeps its own and
	// hides the lifted one.
	var footers int
	for _, n := range screens[3].Nodes {
		if n.Name == "Footer" {
			footers++
		}
	}
	if footers != 1 {
		t.Errorf("page Four has %d footers", footers)
	}
}
