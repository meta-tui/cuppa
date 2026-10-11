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

// prefsScreen is a settings screen: a title, the rows of settings with their
// own column colours, a model view and a key bar of hints the program sets.
func prefsScreen() design.Document {
	doc := design.NewDocument("Prefs", 60, 14)
	doc.Theme.Palette = []design.Swatch{{Name: "Accent", Color: "#ff0000"}, {Name: "Selected", Color: "#00005f"}, {Name: "Focus", Color: "#00f0ff"}}
	doc.Add(design.Node{Component: "lipgloss.rows", Name: "Settings", Rect: design.Rect{X: 1, Y: 1, W: 50, H: 5},
		Props: map[string]string{
			"columns": "Mark:2:fg=@Accent,Label:16:selbg=@Selected,Value:ellipsis:fg=@Focus",
			"rows":    "▸,Parallel,4;,Folder,/very/long/path/that/does/not/fit/at/all/here",
			"styles":  "selected,normal",
			"color":   "212",
		},
		Bind: map[string]string{"rows": "Rows"}})
	doc.Add(design.Node{Component: "lipgloss.slot", Name: "Editor", Rect: design.Rect{X: 1, Y: 7, W: 40, H: 4},
		Props: map[string]string{"view": "sample"}, Bind: map[string]string{"view": "EditorView"}, ShowIf: "Show editor"})
	doc.Add(design.Node{Component: "lipgloss.keybar", Name: "Keys", Rect: design.Rect{X: 0, Y: 13, W: 60, H: 1},
		Props: map[string]string{"hints": "↑↓:move,enter:edit", "keyColor": "220", "labelColor": "240"},
		Bind:  map[string]string{"hints": "Hints"}, Event: "KeyBarClick"})
	return doc
}

func TestKeyBarSlotAndCellStylesBecomeTypedInputs(t *testing.T) {
	p := GenerateScreens([]design.Document{prefsScreen()}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	contract := p.Files["prefs_screen_contract.go"]
	for _, want := range []string{
		"Hints []PrefsHintsHint",
		"type PrefsHintsHint struct{ Key, Label string }",
		`{Key: "↑↓", Label: "move"}`,
		"Rows []PrefsRowsRow",
		"MarkStyle CellStyle",
		"ValueStyle CellStyle",
		"EditorView string",
		"ShowEditor bool",
	} {
		if !strings.Contains(contract, want) {
			t.Errorf("contract lacks %q:\n%s", want, contract)
		}
	}
	if !strings.Contains(p.Files["prefs_screen_view.go"], "func PrefsLayout(") {
		t.Error("no Layout function")
	}
}

const prefsCheck = `package screens

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestPrefs(t *testing.T) {
	p := DefaultPrefsProps()
	f := Prefs(p, 60, 14)
	plain := ansi.Strip(f.View)

	// The key bar: key, label, two spaces between hints.
	lines := strings.Split(plain, "\n")
	if got := strings.TrimRight(lines[13], " "); got != " ↑↓ move  enter edit" {
		t.Errorf("key bar = %q", got)
	}
	p.Hints = []PrefsHintsHint{{Key: "q", Label: "quit"}}
	if got := strings.TrimRight(strings.Split(ansi.Strip(Prefs(p, 60, 14).View), "\n")[13], " "); got != " q quit" {
		t.Errorf("hints from the program = %q", got)
	}

	// A long value is cut with an ellipsis at its column.
	if !strings.Contains(plain, "…") {
		t.Errorf("no ellipsis:\n%s", plain)
	}

	// The accent colour reaches the column that names it, from the palette.
	p = DefaultPrefsProps()
	p.Palette.Accent = "#00ff00"
	if raw := Prefs(p, 60, 14).View; !strings.Contains(raw, "38;2;0;255;0") {
		t.Errorf("the palette colour did not reach the mark column: %q", raw)
	}

	// A cell the program restyles.
	p = DefaultPrefsProps()
	p.Rows[1].ValueStyle = CellStyle{Fg: "#123456", Bold: true}
	if raw := Prefs(p, 60, 14).View; !strings.Contains(raw, "38;2;18;52;86") {
		t.Errorf("the cell style was not used: %q", raw)
	}

	// The selected row gives only its Label column the selected background.
	p = DefaultPrefsProps()
	raw := Prefs(p, 60, 14).View
	if !strings.Contains(raw, "48;2;0;0;95") {
		t.Errorf("no selected background: %q", raw)
	}

	// The model view of the program passes through, colours included.
	p = DefaultPrefsProps()
	p.EditorView = "\x1b[31mred\x1b[0m and text\nsecond line"
	p.ShowEditor = true
	shown := Prefs(p, 60, 14).View
	if !strings.Contains(ansi.Strip(shown), "red and text") || !strings.Contains(shown, "\x1b[31m") {
		t.Errorf("slot view lost or stripped:\n%q", shown)
	}
	// Layout says where it is without drawing.
	r, ok := PrefsLayout(p, 60, 14).RegionNamed("Editor")
	if !ok || r.X != 1 || r.Y != 7 || r.W != 40 || r.H != 4 {
		t.Errorf("layout region = %+v %v", r, ok)
	}
	p.ShowEditor = false
	if _, ok := PrefsLayout(p, 60, 14).RegionNamed("Editor"); ok {
		t.Error("a hidden slot has a region")
	}
	if Prefs(p, 60, 14).View == shown {
		t.Error("hiding the slot changed nothing")
	}

	// Which hint of the key bar a click on the bar is on.
	p = DefaultPrefsProps()
	for x, want := range map[int]int{1: 0, 2: 0, 4: 0, 7: 0, 10: 1, 14: 1} {
		if i, ok := PrefsHintsHintAt(p, x); !ok || i != want {
			t.Errorf("hint at column %d = %d %v, want %d", x, i, ok, want)
		}
	}
	for _, x := range []int{0, 3 + 5, 9, 40} {
		if i, ok := PrefsHintsHintAt(p, x); ok && x != 3 {
			t.Errorf("column %d is on hint %d, want none", x, i)
		}
	}
}
`

func TestBatchOneScreenBuildsAndDraws(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{prefsScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prefs_check_test.go"), []byte(prefsCheck), 0o644); err != nil {
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
