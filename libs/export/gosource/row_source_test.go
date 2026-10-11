package gosource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const rowSourceCheck = `package screens

import (
	"fmt"
	"strings"
	"testing"
)

func rowsOf(n int) []SlotsListRow {
	rows := make([]SlotsListRow, n)
	for i := range rows {
		rows[i] = SlotsListRow{Name: fmt.Sprintf("row %02d", i)}
	}
	return rows
}

func TestTheRowSourceIsAskedForTheSizeOfItsPlace(t *testing.T) {
	all := rowsOf(20)
	p := DefaultSlotsProps()
	var gotW, gotH int
	p.ListRows = func(w, h int) ([]SlotsListRow, int) {
		gotW, gotH = w, h
		return Window(all, 10, h)
	}
	f := Slots(p, 50, 8)
	if gotW != 40 || gotH != 6 {
		t.Fatalf("the source was given %dx%d, want 40x6", gotW, gotH)
	}
	for _, want := range []string{"row 07", "row 12"} {
		if !strings.Contains(f.View, want) {
			t.Errorf("the window lacks %q:\n%s", want, f.View)
		}
	}
	if strings.Contains(f.View, "row 06") || strings.Contains(f.View, "row 13") {
		t.Errorf("the window shows rows outside it:\n%s", f.View)
	}
	for _, c := range []struct {
		y    int
		want int
		ok   bool
	}{{1, 7, true}, {6, 12, true}, {0, 0, false}, {7, 0, false}} {
		got, ok := SlotsListAt(p, 50, 8, c.y)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("At y=%d = %d, %v; want %d, %v", c.y, got, ok, c.want, c.ok)
		}
	}
	// Without a source the list is drawn as before, and At counts from 0.
	p.ListRows = nil
	p.List = all[:3]
	if f := Slots(p, 50, 8); !strings.Contains(f.View, "row 02") {
		t.Errorf("the list is not drawn:\n%s", f.View)
	}
	if got, ok := SlotsListAt(p, 50, 8, 3); !ok || got != 2 {
		t.Errorf("At without a source = %d, %v", got, ok)
	}
	if _, ok := SlotsListAt(p, 50, 8, 4); ok {
		t.Error("At found a row below the last")
	}
}

func TestWindowKeepsTheSelectedRowCentredAndPinnedAtTheEnds(t *testing.T) {
	rows := rowsOf(10)
	for _, c := range []struct{ sel, h, first, n int }{
		{0, 4, 0, 4}, {1, 4, 0, 4}, {5, 4, 3, 4}, {9, 4, 6, 4}, {-3, 4, 0, 4}, {99, 4, 6, 4},
		{5, 0, 0, 10}, {5, 10, 0, 10}, {5, 12, 0, 10},
	} {
		got, first := Window(rows, c.sel, c.h)
		if first != c.first || len(got) != c.n {
			t.Errorf("Window(sel %d, h %d) = first %d, %d rows; want %d, %d", c.sel, c.h, first, len(got), c.first, c.n)
		}
	}
}
`

// The rows of a list can come from a function that is given the size the
// layout gave them (ADR: no two-pass Layout in the program), with a Window
// helper and an At that maps a click to a row.
func TestRowsComeFromAFunctionOfTheSizeOfTheirPlace(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{slotsScreen()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "row_source_check_test.go"), []byte(rowSourceCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	goRun(t, dir, "mod", "tidy")
	goRun(t, dir, "vet", "./...")
	goRun(t, dir, "test", "./...")
}

// The slice layout writes the same function and At, and the runtime's Window.
func TestTheSliceLayoutHasTheRowSource(t *testing.T) {
	p, err := GenerateSlices(slotsScreen(), standard.Default(), Options{Module: "example.com/slots"})
	if err != nil {
		t.Fatal(err)
	}
	page := strings.Join(strings.Fields(p.Files["pages/slots/slots.cuppa.go"]), " ")
	for _, want := range []string{"ListRows func(w, h int) ([]ListRow, int)", "func ListAt(p Props, w, h, y int) (int, bool)"} {
		if !strings.Contains(page, want) {
			t.Errorf("the page lacks %q", want)
		}
	}
	if !strings.Contains(p.Files["internal/cuppa/runtime.go"]+p.Files["internal/cuppa/screens.go"], "func Window[T any]") {
		t.Error("the runtime has no Window")
	}
}
