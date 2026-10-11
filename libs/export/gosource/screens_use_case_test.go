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

// coloursScreen is a small screen with a bound title, a list, a status line
// that can be hidden, a button that raises an event and keys.
func coloursScreen() design.Document {
	doc := design.NewDocument("Colours", 60, 12)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 2, Y: 0, W: 30, H: 1},
		Props: map[string]string{"text": "Colours", "color": "@Accent"}, Bind: map[string]string{"text": "Title"}})
	doc.Add(design.Node{Component: "lipgloss.list", Name: "Slots", Rect: design.Rect{X: 2, Y: 2, W: 30, H: 5},
		Props: map[string]string{"items": "Accent, Focus"}, Bind: map[string]string{"items": "Slots"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Status", Rect: design.Rect{X: 2, Y: 8, W: 30, H: 1},
		Props: map[string]string{"text": "saved"}, Bind: map[string]string{"text": "Status"}, ShowIf: "Show status"})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Save button", Rect: design.Rect{X: 40, Y: 2, W: 12, H: 3},
		Props: map[string]string{"title": "Save"}, Event: "Save", Layout: design.Layout{X: "100% - 20"}})
	doc.Keys = []design.KeyBinding{{Key: "s", Event: "Save", Label: "save"}, {Key: "esc", Event: "Back", Label: "back"}}
	doc.Theme.Palette = []design.Swatch{{Name: "Accent", Color: "#ff0000"}, {Name: "Dim colour", Color: "#666666"}}
	return doc
}

func TestScreensExportContract(t *testing.T) {
	p := GenerateScreens([]design.Document{coloursScreen()}, standard.Default(), "screens")
	contract := p.Files["colours_screen_contract.go"]
	for _, want := range []string{
		"type ColoursProps struct",
		"Title string",
		"Slots []string",
		"Status string",
		"ShowStatus bool",
		"Theme Theme",
		"Palette ColoursPalette",
		"type ColoursPalette struct",
		"Accent string",
		"DimColour string",
		"func DefaultColoursPalette() ColoursPalette",
		"func DefaultColoursProps() ColoursProps",
		`[]string{"Accent", "Focus"}`,
		"type ColoursSave struct{ X, Y int }",
		"type ColoursBack struct{ X, Y int }",
		`{Key: "esc", Label: "back", Event: "Back"}`,
	} {
		if !strings.Contains(contract, want) {
			t.Errorf("contract lacks %q:\n%s", want, contract)
		}
	}
	for _, name := range []string{"colours_screen_view.go", "cuppa_runtime.go", "cuppa_widgets.go", "cuppa_screens.go"} {
		if !strings.HasPrefix(p.Files[name], GeneratedHeader) {
			t.Errorf("%s is missing or lacks the generated header", name)
		}
	}
	if len(p.Notes) != 0 {
		t.Errorf("notes: %v", p.Notes)
	}
	if _, ok := p.Files["main.go"]; ok {
		t.Error("a screens package must not have a main.go")
	}
}

func TestScreensExportReportsConflicts(t *testing.T) {
	doc := design.NewDocument("Clash", 40, 6)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "A", Rect: design.Rect{W: 10, H: 1},
		Props: map[string]string{"text": "a"}, Bind: map[string]string{"text": "Value"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "B", Rect: design.Rect{Y: 2, W: 10, H: 3},
		Bind: map[string]string{"border": "Value"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "C", Rect: design.Rect{Y: 5, W: 10, H: 1},
		Props: map[string]string{"bold": "true"}, Bind: map[string]string{"bold": "Value"}})
	p := GenerateScreens([]design.Document{doc, doc}, standard.Default(), "screens")
	joined := strings.Join(p.Notes, "\n")
	if !strings.Contains(joined, "no free identifier") {
		t.Errorf("duplicate design name not reported: %v", p.Notes)
	}
	if !strings.Contains(joined, `input "Value" is used as`) {
		t.Errorf("type clash not reported: %v", p.Notes)
	}
}

func TestWriteScreensKeepsHandWrittenFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "colours_screen_view.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := GenerateScreens([]design.Document{coloursScreen()}, standard.Default(), "screens")
	if err := WriteScreens(dir, p); err == nil {
		t.Fatal("a hand-written file was overwritten")
	}
	_ = os.Remove(filepath.Join(dir, "colours_screen_view.go"))
	if err := WriteScreens(dir, p); err != nil {
		t.Fatal(err)
	}
	if err := WriteScreens(dir, p); err != nil {
		t.Fatalf("second export must replace generated files: %v", err)
	}
	// Writing one screen keeps the others; replacing the folder drops the
	// screens that are gone.
	other := coloursScreen()
	other.Name = "Settings"
	if err := ReplaceScreens(dir, GenerateScreens([]design.Document{other}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "colours_screen_view.go")); err == nil {
		t.Error("stale generated file kept")
	}
	if _, err := os.Stat(filepath.Join(dir, "settings_screen_view.go")); err != nil {
		t.Error("new screen missing")
	}
}

const coloursCheck = `package screens

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestColoursScreen(t *testing.T) {
	p := DefaultColoursProps()
	p.Title = "Palette"
	p.Slots = []string{"Hot", "Cold, with a comma", "Mild"}
	f := Colours(p, 60, 12)
	for _, want := range []string{"Palette", "Hot", "Cold, with a comma", "Mild", "saved"} {
		if !strings.Contains(f.View, want) {
			t.Errorf("view lacks %q:\n%s", want, f.View)
		}
	}
	if w, h := ColoursSize(); w != 60 || h != 12 {
		t.Errorf("size %dx%d", w, h)
	}

	p.ShowStatus = false
	if got := Colours(p, 60, 12).View; strings.Contains(got, "saved") {
		t.Errorf("hidden status still drawn:\n%s", got)
	}

	// The button follows the width it is drawn at: 100% - 20.
	wide := Colours(p, 100, 12)
	r, ok := wide.RegionNamed("Save button")
	if !ok || r.X != 80 {
		t.Fatalf("button region at width 100: %+v %v", r, ok)
	}
	ev, ok := ColoursHandle(wide, tea.MouseClickMsg{X: 82, Y: 3, Button: tea.MouseLeft})
	save, isSave := ev.(ColoursSave)
	if !ok || !isSave || save.X != 2 || save.Y != 1 {
		t.Errorf("click gave %#v %v", ev, ok)
	}
	if _, ok := ColoursHandle(wide, tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft}); ok {
		t.Error("a click elsewhere raised an event")
	}
	ev, ok = ColoursHandle(wide, tea.KeyPressMsg{Code: tea.KeyEscape})
	if _, isBack := ev.(ColoursBack); !ok || !isBack {
		t.Errorf("esc gave %#v %v", ev, ok)
	}
	ev, ok = ColoursHandle(wide, tea.KeyPressMsg{Code: 's', Text: "s"})
	if _, isSave := ev.(ColoursSave); !ok || !isSave {
		t.Errorf("s gave %#v %v", ev, ok)
	}

	// A theme replaces the colours components follow.
	plain := Colours(DefaultColoursProps(), 60, 12).View
	p2 := DefaultColoursProps()
	p2.Theme = Theme{Text: "#ff0000", Background: "#001122"}
	themed := Colours(p2, 60, 12).View
	if plain == themed {
		t.Errorf("theme did not change the drawing:\n%q", themed)
	}
	// A palette colour changed by the program reaches every component that
	// uses it, and an empty one keeps the design's colour.
	red := Colours(DefaultColoursProps(), 60, 12).View
	p3 := DefaultColoursProps()
	p3.Palette.Accent = "#00ff00"
	green := Colours(p3, 60, 12).View
	if !strings.Contains(red, "255;0;0") || !strings.Contains(green, "0;255;0") || strings.Contains(green, "255;0;0") {
		t.Errorf("palette colour did not follow: red %q green %q", red, green)
	}
	p3.Palette.Accent = ""
	if again := Colours(p3, 60, 12).View; again != red {
		t.Error("an empty palette colour must keep the design's")
	}
	if len(ColoursKeys) != 2 {
		t.Errorf("keys %v", ColoursKeys)
	}
}
`

// TestTheGeneratedScreensBuildAndWork builds the package and drives a screen:
// bound values reach the drawing, a hidden component disappears, a click and a
// key come back as events and a theme replaces the design's colours.
func TestTheGeneratedScreensBuildAndWork(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{coloursScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "colours_check_test.go"), []byte(coloursCheck), 0o644); err != nil {
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
