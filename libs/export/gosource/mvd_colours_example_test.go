package gosource

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const mvdColoursExample = "../../../examples/mvd-colours.cuppa"

// mvdSlots are the colours MVD lets the person change, with the value each
// has by default.
var mvdSlots = []struct{ name, label, color string }{
	{"Accent", "Accent (title)", "#ff007f"},
	{"Focus", "Focus (active border)", "#00f0ff"},
	{"Highlight", "Highlight (keys, bars)", "#ffe600"},
	{"Success", "Success", "#3ddc84"},
	{"Error", "Error", "#ff4d4d"},
	{"Dim", "Dim (borders, hints)", "#6b7280"},
	{"Text", "Text", "#e5e7eb"},
	{"Selected", "Selected row (background)", "#3b0f2a"},
}

// mvdColours builds, through the editor, the design of MVD's Colours screen
// the way it would be recreated in Cuppa: the interface colours are the
// design's named colours, so the program can change them, and each slot is a
// colour swatch that uses its own name.
func mvdColours(t *testing.T) design.Document {
	t.Helper()
	ed := editor.New(standard.Default(), design.NewDocument("MVD Colours", 80, 24))
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	add := func(comp, name string, x, y, w, h int) design.NodeID {
		t.Helper()
		id, err := ed.Add(comp, x, y)
		must(err)
		ed.SetRect(id, design.Rect{X: x, Y: y, W: w, H: h}, false)
		ed.Rename(id, name)
		return id
	}
	use := func(id design.NodeID, key, swatch string) { must(ed.UseSwatch(id, key, swatch)) }

	for _, s := range mvdSlots {
		must(ed.SetSwatch(s.name, s.color))
	}

	frame := add("lipgloss.box", "Frame", 0, 0, 80, 24)
	must(ed.SetProp(frame, "title", "MVD · Colours"))
	use(frame, "color", "Dim")
	must(ed.SetLayout(frame, editor.AxisW, "100%"))
	must(ed.SetLayout(frame, editor.AxisH, "100%"))

	for i, s := range mvdSlots {
		marker := add("lipgloss.label", s.name+" cursor", 2, 2+i, 1, 1)
		must(ed.SetProp(marker, "text", "▸"))
		use(marker, "color", "Accent")
		must(ed.SetShowIf(marker, "Cursor on "+s.name))

		row := add("lipgloss.swatch", s.name, 4, 2+i, 70, 1)
		must(ed.SetProp(row, "label", s.label))
		must(ed.SetProp(row, "labelWidth", "28"))
		use(row, "color", s.name)
		use(row, "textColor", "Text")
		must(ed.SetLayout(row, editor.AxisW, "100% - 6"))
		must(ed.SetBinding(row, "label", s.name+" label"))
		must(ed.SetEvent(row, "Pick "+s.name))
	}

	help := add("lipgloss.label", "Help", 2, 12, 76, 1)
	must(ed.SetProp(help, "text", "Colours are #rgb or #rrggbb and apply as soon as you accept them."))
	use(help, "color", "Dim")
	must(ed.SetLayout(help, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(help, "text", "Help"))

	status := add("lipgloss.label", "Status", 2, 14, 76, 1)
	must(ed.SetProp(status, "text", "not a colour"))
	use(status, "color", "Error")
	must(ed.SetLayout(status, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(status, "text", "Status"))
	must(ed.SetShowIf(status, "Show status"))

	keys := add("lipgloss.label", "Key bar", 2, 22, 76, 1)
	must(ed.SetProp(keys, "text", "↑↓ move · enter edit · d default · s save · esc back"))
	use(keys, "color", "Dim")
	must(ed.SetLayout(keys, editor.AxisY, "100% - 2"))
	must(ed.SetLayout(keys, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(keys, "text", "Key bar"))

	must(ed.SetKeys("enter=Edit:edit, d=Default:default, s=Save:save, esc=Back:back"))
	return ed.Document()
}

// TestMVDColoursExample keeps examples/mvd-colours.cuppa in step with the
// design above (set CUPPA_WRITE_EXAMPLES=1 to write it) and checks the screen
// it exports: the contract MVD's Colours screen is written against, with the
// interface colours as typed palette fields.
func TestMVDColoursExample(t *testing.T) {
	want := mvdColours(t)
	path := filepath.FromSlash(mvdColoursExample)
	if os.Getenv("CUPPA_WRITE_EXAMPLES") == "1" {
		if _, err := disk.Save(path, want); err != nil {
			t.Fatal(err)
		}
	}
	got, err := disk.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("examples/mvd-colours.cuppa is out of date; run with CUPPA_WRITE_EXAMPLES=1\n got %+v\nwant %+v", got, want)
	}

	p := GenerateScreens([]design.Document{got}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	contract := p.Files["mvd_colours_screen_contract.go"]
	for _, wantSrc := range []string{
		"type MVDColoursProps struct",
		"Palette MVDColoursPalette",
		"type MVDColoursPalette struct",
		"Accent string",
		"Selected string",
		"CursorOnAccent bool",
		"AccentLabel string",
		"Status string",
		"ShowStatus bool",
		"KeyBar string",
		"type MVDColoursPickAccent struct{ X, Y int }",
		"type MVDColoursPickSelected struct{ X, Y int }",
		"type MVDColoursEdit struct{ X, Y int }",
		"type MVDColoursSave struct{ X, Y int }",
		"type MVDColoursBack struct{ X, Y int }",
	} {
		if !strings.Contains(contract, wantSrc) {
			t.Errorf("contract lacks %q:\n%s", wantSrc, contract)
		}
	}
}
