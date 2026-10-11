package gosource

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// paneScreen is a download-like screen: a box of playlists whose height
// follows how many there are (within a third of the window), the entries box
// under it taking the rest, and an output box that only exists in a wide
// window.
func paneScreen() design.Document {
	doc := design.NewDocument("Pane", 120, 36)
	doc.Add(design.Node{Component: "bubbles.paginator", Name: "Counter", Rect: design.Rect{X: 0, Y: 0, W: 10, H: 1},
		Props: map[string]string{"total": "3"}, Bind: map[string]string{"total": "PlaylistCount"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Playlists", Rect: design.Rect{X: 0, Y: 3, W: 40, H: 5},
		Props:  map[string]string{"title": "Playlists"},
		Layout: design.Layout{W: "38%", H: "min(max($PlaylistCount + 2, 5), (100% - 4) / 3)"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Entries", Rect: design.Rect{X: 0, Y: 8, W: 40, H: 27},
		Props:  map[string]string{"title": "Entries"},
		Layout: design.Layout{W: "38%", Y: `below("Playlists")`, H: `100% - below("Playlists") - 1`}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Output", Rect: design.Rect{X: 40, Y: 3, W: 80, H: 32},
		Props:  map[string]string{"title": "Output"},
		ShowIf: "w >= 100 && !$FullLog",
		Layout: design.Layout{X: `right("Playlists")`, W: "100% - 40", H: "100% - 4"}})
	return doc
}

func TestLayoutReadsInputsAndTheComponentsBefore(t *testing.T) {
	p := GenerateScreens([]design.Document{paneScreen()}, standard.Default(), "screens")
	joined := strings.Join(p.Notes, "\n")
	if !strings.Contains(joined, `"FullLog"`) {
		t.Errorf("an input read by a condition and bound to nothing is not reported: %v", p.Notes)
	}
	contract := p.Files["pane_screen_contract.go"]
	for _, want := range []string{"PlaylistCount int", "FullLog int"} {
		if !strings.Contains(contract, want) {
			t.Errorf("contract lacks %q:\n%s", want, contract)
		}
	}
	view := p.Files["pane_screen_view.go"]
	for _, want := range []string{"func paneEnv(", "FitEnv: func(e fitEnv)", `e.In("PlaylistCount")`, `e.Rect("Playlists", "bottom")`, "func PaneLayout("} {
		if !strings.Contains(view, want) {
			t.Errorf("view lacks %q", want)
		}
	}
}

const paneCheck = `package screens

import "testing"

func TestPane(t *testing.T) {
	p := DefaultPaneProps()
	region := func(f Frame, name string) (Region, bool) { return f.RegionNamed(name) }

	// Three playlists: five rows (the minimum), the entries under them.
	f := PaneLayout(p, 120, 36)
	pl, _ := region(f, "Playlists")
	en, _ := region(f, "Entries")
	if pl.H != 5 || pl.W != 45 || en.Y != pl.Y+pl.H || en.H != 36-(pl.Y+pl.H)-1 {
		t.Errorf("playlists %+v entries %+v", pl, en)
	}
	out, ok := region(f, "Output")
	if !ok || out.X != pl.X+pl.W || out.W != 80 {
		t.Errorf("output %+v %v", out, ok)
	}

	// Thirty playlists: a third of the window at most, and the entries follow.
	p.PlaylistCount = 30
	f = PaneLayout(p, 120, 36)
	pl, _ = region(f, "Playlists")
	en, _ = region(f, "Entries")
	if pl.H != (36-4)/3 || en.Y != pl.Y+pl.H {
		t.Errorf("30 playlists: %+v %+v", pl, en)
	}

	// A narrow window has no output box; a full log has none either.
	if _, ok := region(PaneLayout(p, 80, 24), "Output"); ok {
		t.Error("output shown in a narrow window")
	}
	p.FullLog = 1
	if _, ok := region(PaneLayout(p, 120, 36), "Output"); ok {
		t.Error("output shown in the full log")
	}

	// The drawing places the same rectangles.
	g := Pane(DefaultPaneProps(), 120, 36)
	if r, _ := g.RegionNamed("Entries"); r.Y != 8 {
		t.Errorf("drawn entries at %d", r.Y)
	}
}
`

func TestPaneLayoutBuildsAndFollowsTheInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{paneScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pane_check_test.go"), []byte(paneCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"mod", "tidy"}, {"vet", "./..."}, {"test", "./..."}} {
		cmd := exec.Command(goBin, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}
