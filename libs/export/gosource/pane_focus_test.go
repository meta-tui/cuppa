package gosource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// focusPaneScreen is a pane: a box and its rows, with one focus switch for both.
func focusPaneScreen() design.Document {
	doc := design.NewDocument("Pane", 40, 8)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 40, H: 8},
		Props: map[string]string{"title": "Entries", "color": "212", "blurColor": "240"},
		Bind:  map[string]string{"focused": "Focus"}})
	doc.Add(design.Node{Component: "lipgloss.rows", Name: "Entries", Rect: design.Rect{X: 1, Y: 1, W: 38, H: 6},
		Props: map[string]string{"columns": "Name:20:fg=#e5e7eb,Size", "rows": "one,1;two,2", "styles": "selected,normal"},
		Bind:  map[string]string{"rows": "List", "focused": "Focus"}})
	return doc
}

const focusPaneCheck = `package screens

import (
	"strings"
	"testing"
)

func TestAFocusPane(t *testing.T) {
	p := DefaultPaneProps()
	p.Focus = true
	on := Pane(p, 40, 8).View
	if !strings.Contains(on, "48;5;212") {
		t.Errorf("a focused pane does not invert its selected row:\n%q", on)
	}
	p.Focus = false
	off := Pane(p, 40, 8).View
	if strings.Contains(off, "48;5;212") {
		t.Errorf("a pane without focus inverts its selected row:\n%q", off)
	}
	if !strings.Contains(off, "38;5;240") {
		t.Errorf("a pane without focus keeps the focused border:\n%q", off)
	}
	if !strings.Contains(off, "38;5;212") {
		t.Errorf("the selected row is not accent text:\n%q", off)
	}
}
`

func TestAPaneWithoutFocusDrawsAccentTextAndTheBlurredBorder(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{focusPaneScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "focus_pane_check_test.go"), []byte(focusPaneCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	goRun(t, dir, "mod", "tidy")
	goRun(t, dir, "vet", "./...")
	goRun(t, dir, "test", "./...")
}
