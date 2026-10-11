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

// slotsScreen is a screen whose body is one bound Rows component: the program
// gives the rows and the style of each.
func slotsScreen() design.Document {
	doc := design.NewDocument("Slots", 50, 8)
	doc.Add(design.Node{Component: "lipgloss.rows", Name: "Slot rows", Rect: design.Rect{X: 1, Y: 1, W: 40, H: 6},
		Props: map[string]string{
			"columns": "Mark:2,Name:14,Swatch:6:colour,Hex",
			"rows":    "▸,Accent,#ff007f,#ff007f;,Focus,#00f0ff,#00f0ff",
			"styles":  "selected,normal",
		},
		Bind:  map[string]string{"rows": "List"},
		Event: "Pick"})
	return doc
}

func TestRowsBecomeATypedListInTheContract(t *testing.T) {
	p := GenerateScreens([]design.Document{slotsScreen()}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	contract := p.Files["slots_screen_contract.go"]
	for _, want := range []string{
		"List []SlotsListRow",
		"type SlotsListRow struct",
		"Mark string",
		"Name string",
		"Swatch string",
		"Hex string",
		"Style RowStyle",
		`{Mark: "▸", Name: "Accent", Swatch: "#ff007f", Hex: "#ff007f", Style: RowSelected}`,
		`{Mark: "", Name: "Focus", Swatch: "#00f0ff", Hex: "#00f0ff", Style: RowNormal}`,
	} {
		if !strings.Contains(contract, want) {
			t.Errorf("contract lacks %q:\n%s", want, contract)
		}
	}
	if !strings.Contains(p.Files["cuppa_screens.go"], "RowSelected") {
		t.Error("the shared file lacks the row styles")
	}
}

func TestStylesBoundAlongsideRowsAreIgnoredWithANote(t *testing.T) {
	doc := slotsScreen()
	doc.Nodes[0].Bind["styles"] = "Looks"
	p := GenerateScreens([]design.Document{doc}, standard.Default(), "screens")
	if !strings.Contains(strings.Join(p.Notes, "\n"), "follow the Style of each row") {
		t.Errorf("no note: %v", p.Notes)
	}
	if strings.Contains(p.Files["slots_screen_contract.go"], "Looks") {
		t.Error("the styles binding became an input")
	}
}

const slotsCheck = `package screens

import (
	"strings"
	"testing"
)

func TestSlotsRows(t *testing.T) {
	p := DefaultSlotsProps()
	if len(p.List) != 2 || p.List[0].Style != RowSelected {
		t.Fatalf("default rows %+v", p.List)
	}
	p.List = []SlotsListRow{
		{Mark: "▸", Name: "Hot, sweet", Swatch: "#ff0000", Hex: "#ff0000", Style: RowSelected},
		{Name: "  indented", Style: RowDim},
		{Name: "plain"},
	}
	f := Slots(p, 50, 8)
	for _, want := range []string{"Hot, sweet", "  indented", "plain", "#ff0000"} {
		if !strings.Contains(f.View, want) {
			t.Errorf("view lacks %q:\n%s", want, f.View)
		}
	}
	if !strings.Contains(f.View, "48;5;212") {
		t.Errorf("the selected row is not drawn in the accent colour:\n%q", f.View)
	}
	p.List[0].Style = RowNormal
	if strings.Contains(Slots(p, 50, 8).View, "48;5;212") {
		t.Error("no row is selected, but one is still drawn inverted")
	}
	p.List = nil
	if strings.Contains(Slots(p, 50, 8).View, "plain") {
		t.Error("an empty list still draws rows")
	}
}
`

// TestRowsScreenBuildsAndDraws builds the package and checks that typed rows
// reach the drawing: cells keep their commas and spaces, the style of a row
// decides how it is drawn, and an empty list draws nothing.
func TestRowsScreenBuildsAndDraws(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{slotsScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "slots_check_test.go"), []byte(slotsCheck), 0o644); err != nil {
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
