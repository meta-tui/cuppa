package gosource

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	colourpicker "github.com/meta-ui-labs/bubble-colourpicker"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

const libraryModule = "github.com/meta-ui-labs/bubble-colourpicker"

// libraryFile reads a file of the published module (the version in go.mod).
func libraryFile(t *testing.T, name string) string {
	t.Helper()
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", libraryModule).Output()
	if err != nil {
		t.Skipf("the module is not available: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(out)), name))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// The program carries the library's code, not a second implementation: the
// embedded copy is the module's file byte for byte.
func TestColourPickerCopyIsTheLibrarysFile(t *testing.T) {
	if testing.Short() {
		t.Skip("reads the module with the go tool")
	}
	want := libraryFile(t, "colourpicker.go")
	got := strings.ReplaceAll(colourPickerLibrary, "\r\n", "\n")
	if got != want {
		t.Fatal("colourpicker_go.txt is not colourpicker.go of " + libraryModule + " " + colourPickerVersion +
			"; copy it from the module (go list -m -f '{{.Dir}}' " + libraryModule + ") and set colourPickerVersion")
	}
}

func TestColourPickerSourceIsRenamedForAProgram(t *testing.T) {
	src := colourPickerSource()
	for _, want := range []string{
		"package main",
		"type ColourPicker struct",
		"func NewColourPicker(",
		"func ColourPickerWithValue(",
		"type ColourPickerChangedMsg struct",
		"ColourPickerTab256",
		`extra["lipgloss.colourpicker"] = colourPickerPart`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("source lacks %q", want)
		}
	}
	for _, bare := range []string{"type Model ", "func New(", "func WithValue(", "type ChangedMsg ", "package colourpicker"} {
		if strings.Contains(src, bare) {
			t.Errorf("source still has %q", bare)
		}
	}
}

// The designer's picker and the library's draw the same characters, for every
// tab and for a palette colour, a hex colour and none.
func TestDesignerPickerDrawsWhatTheLibraryDraws(t *testing.T) {
	cat := standard.Default()
	for _, tab := range []string{"16", "256", "RGB", "HSL"} {
		for _, value := range []string{"", "212", "9", "#ff007f", "#808080", "#00ff00"} {
			for _, slide := range []int{0, 1, 2} {
				node := design.Node{Component: "lipgloss.colourpicker", Rect: design.Rect{W: 46, H: 26},
					Props: map[string]string{"value": value, "tab": tab, "slide": string(rune('0' + slide)), "tabs": "16,256,RGB,HSL"}}
				got := trimRows(strings.Split(ansi.Strip(strings.Join(scene.RenderNode(node, cat).Lines(), "\n")), "\n"), 26)
				m := colourpicker.New(colourpicker.WithValue(value), colourpicker.WithTab(tab), colourpicker.WithAccent("212")).SetSlide(slide)
				want := trimRows(strings.Split(ansi.Strip(m.View()), "\n"), 26)
				if strings.Join(got, "\n") != strings.Join(want, "\n") {
					t.Fatalf("tab %s, value %q, slider %d differ\n-- designer\n%s\n-- library\n%s", tab, value, slide, strings.Join(got, "\n"), strings.Join(want, "\n"))
				}
			}
		}
	}
	// A shorter list of tabs.
	node := design.Node{Component: "lipgloss.colourpicker", Rect: design.Rect{W: 46, H: 12},
		Props: map[string]string{"value": "#123456", "tab": "HSL", "tabs": "hsl, 16"}}
	got := trimRows(strings.Split(ansi.Strip(strings.Join(scene.RenderNode(node, cat).Lines(), "\n")), "\n"), 12)
	want := trimRows(strings.Split(ansi.Strip(colourpicker.New(colourpicker.WithValue("#123456"), colourpicker.WithTabs("hsl", "16"), colourpicker.WithTab("HSL")).View()), "\n"), 12)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("two tabs differ\n-- designer\n%s\n-- library\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func pickerScreen() design.Document {
	doc := design.NewDocument("Palette", 60, 14)
	doc.Add(design.Node{Component: "lipgloss.colourpicker", Name: "Picker", Rect: design.Rect{X: 2, Y: 1, W: 46, H: 10},
		Props: map[string]string{"value": "#ff007f", "tab": "RGB", "tabs": "16,RGB,HSL"},
		Bind:  map[string]string{"value": "Colour"}, ShowIf: "Show picker"})
	return doc
}

func TestColourPickerBecomesATypedInput(t *testing.T) {
	p := GenerateScreens([]design.Document{pickerScreen()}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	contract := p.Files["palette_screen_contract.go"]
	for _, want := range []string{
		"Colour ColourPicker",
		`NewColourPicker(ColourPickerWithValue("#ff007f"), ColourPickerWithTabs("16", "RGB", "HSL"), ColourPickerWithTab("RGB")`,
		"ShowPicker bool",
	} {
		if !strings.Contains(contract, want) {
			t.Errorf("contract lacks %q:\n%s", want, contract)
		}
	}
	if !strings.Contains(p.Files["cuppa_colourpicker.go"], "type ColourPicker struct") {
		t.Error("the package lacks the picker")
	}
}

const pickerCheck = `package screens

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPaletteScreen(t *testing.T) {
	p := DefaultPaletteProps()
	if p.Colour.Value() != "#ff007f" || p.Colour.Tab() != "RGB" {
		t.Fatalf("default picker: %q on %q", p.Colour.Value(), p.Colour.Tab())
	}
	f := Palette(p, 60, 14)
	if !strings.Contains(ansi.Strip(f.View), "Value #ff007f") {
		t.Fatalf("the screen does not show the picker:\n%s", ansi.Strip(f.View))
	}
	r, ok := f.RegionNamed("Picker")
	if !ok {
		t.Fatal("no region for the picker")
	}
	// The program owns the picker: it gives it its position and its messages.
	p.Colour = p.Colour.SetOrigin(r.X, r.Y)
	var changed bool
	p.Colour, _ = p.Colour.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	changed = p.Colour.Value() == "#fe007f"
	if !changed {
		t.Errorf("the left arrow did not lower the red channel: %q", p.Colour.Value())
	}
	p.Colour, _ = p.Colour.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if p.Colour.Tab() != "HSL" {
		t.Errorf("tab key = %q", p.Colour.Tab())
	}
	again := ansi.Strip(Palette(p, 60, 14).View)
	if !strings.Contains(again, "Value "+p.Colour.Value()) || strings.Contains(again, "Value #ff007f") {
		t.Errorf("the screen did not follow the picker:\n%s", again)
	}
	p.ShowPicker = false
	if strings.Contains(ansi.Strip(Palette(p, 60, 14).View), "Value ") {
		t.Error("a hidden picker is still drawn")
	}
}
`

func TestColourPickerScreenBuildsAndFollowsTheProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{pickerScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "palette_check_test.go"), []byte(pickerCheck), 0o644); err != nil {
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
