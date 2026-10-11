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

const mvdColoursRowsExample = "../../../examples/mvd-colours-rows.cuppa"

// mvdColoursRows is MVD's Colours screen again, with one Rows component for
// the eight slots instead of a swatch and a cursor marker per slot (#176): the
// program gives the rows and marks the selected one with a row style.
func mvdColoursRows(t *testing.T) design.Document {
	t.Helper()
	ed := editor.New(standard.Default(), design.NewDocument("MVD Colours Rows", 80, 24))
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
	for _, s := range mvdSlots {
		must(ed.SetSwatch(s.name, s.color))
	}

	frame := add("lipgloss.box", "Frame", 0, 0, 80, 24)
	must(ed.SetProp(frame, "title", "MVD · Colours"))
	must(ed.UseSwatch(frame, "color", "Dim"))
	must(ed.SetLayout(frame, editor.AxisW, "100%"))
	must(ed.SetLayout(frame, editor.AxisH, "100%"))

	// A cell with a comma in it (the labels have some) is escaped with a backslash.
	esc := strings.NewReplacer(string(rune(92)), string(rune(92))+string(rune(92)), ",", string(rune(92))+",", ";", string(rune(92))+";")
	var rows, styles []string
	for i, s := range mvdSlots {
		mark, style := "", "normal"
		if i == 0 {
			mark, style = "▸", "selected"
		}
		rows = append(rows, strings.Join([]string{mark, esc.Replace(s.label), s.color, s.color}, ","))
		styles = append(styles, style)
	}
	list := add("lipgloss.rows", "Slots", 2, 2, 76, len(mvdSlots))
	must(ed.SetProp(list, "columns", "Mark:2,Name:28,Swatch:6:colour,Value"))
	must(ed.SetProp(list, "rows", strings.Join(rows, ";")))
	must(ed.SetProp(list, "styles", strings.Join(styles, ",")))
	must(ed.UseSwatch(list, "color", "Accent"))
	must(ed.SetLayout(list, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(list, "rows", "Slots"))
	must(ed.SetEvent(list, "Pick"))

	help := add("lipgloss.label", "Help", 2, 10, 76, 1)
	must(ed.SetProp(help, "text", "Colours are #rgb or #rrggbb and apply as soon as you accept them."))
	must(ed.UseSwatch(help, "color", "Dim"))
	must(ed.SetLayout(help, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(help, "text", "Help"))

	// The colour picker opens under the list on the slot that is being
	// changed: the program owns it and shows it with "Show picker".
	picker := add("lipgloss.colourpicker", "Picker", 2, 11, 46, 10)
	must(ed.SetProp(picker, "tabs", "16,RGB,HSL"))
	must(ed.SetProp(picker, "tab", "RGB"))
	must(ed.UseSwatch(picker, "color", "Accent"))
	must(ed.SetBinding(picker, "value", "Colour"))
	must(ed.SetShowIf(picker, "Show picker"))

	status := add("lipgloss.label", "Status", 2, 21, 76, 1)
	must(ed.SetProp(status, "text", "not a colour"))
	must(ed.UseSwatch(status, "color", "Error"))
	must(ed.SetLayout(status, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(status, "text", "Status"))
	must(ed.SetShowIf(status, "Show status"))

	keys := add("lipgloss.label", "Key bar", 2, 22, 76, 1)
	must(ed.SetProp(keys, "text", "↑↓ move · enter edit · d default · s save · esc back"))
	must(ed.UseSwatch(keys, "color", "Dim"))
	must(ed.SetLayout(keys, editor.AxisY, "100% - 2"))
	must(ed.SetLayout(keys, editor.AxisW, "100% - 4"))
	must(ed.SetBinding(keys, "text", "Key bar"))

	must(ed.SetKeys("enter=Edit:edit, d=Default:default, s=Save:save, esc=Back:back"))
	return ed.Document()
}

// TestMVDColoursRowsExample keeps examples/mvd-colours-rows.cuppa in step with
// the design above (set CUPPA_WRITE_EXAMPLES=1 to write it) and checks the
// contract it exports: one typed list for the eight slots.
func TestMVDColoursRowsExample(t *testing.T) {
	want := mvdColoursRows(t)
	path := filepath.FromSlash(mvdColoursRowsExample)
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
		t.Fatalf("examples/mvd-colours-rows.cuppa is out of date; run with CUPPA_WRITE_EXAMPLES=1")
	}
	p := GenerateScreens([]design.Document{got}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	contract := p.Files["mvd_colours_rows_screen_contract.go"]
	for _, wantSrc := range []string{
		"Slots []MVDColoursRowsSlotsRow",
		"Colour ColourPicker",
		"ShowPicker bool",
		"type MVDColoursRowsSlotsRow struct",
		"Swatch string",
		"Style RowStyle",
		"type MVDColoursRowsPick struct{ X, Y int }",
		"Palette MVDColoursRowsPalette",
	} {
		if !strings.Contains(contract, wantSrc) {
			t.Errorf("contract lacks %q:\n%s", wantSrc, contract)
		}
	}
	// Eight slots became one list: none of the per-slot inputs is left.
	for _, gone := range []string{"CursorOnAccent", "AccentLabel"} {
		if strings.Contains(contract, gone) {
			t.Errorf("contract still has %s", gone)
		}
	}
}
