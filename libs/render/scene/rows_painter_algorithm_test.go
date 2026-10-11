package scene

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func rowsGrid(w, h int, props map[string]string) *grid.Grid {
	n := design.Node{Component: "lipgloss.rows", Rect: design.Rect{W: w, H: h}, Props: props}
	return RenderNode(n, standard.Default())
}

func TestRowsLayTheCellsOutInTheNamedColumns(t *testing.T) {
	g := rowsGrid(30, 3, map[string]string{
		"columns": "Mark:2,Name:8,Value",
		"rows":    "▸,Accent,#ff007f;,Focus,#00f0ff",
		"styles":  "selected,normal",
	})
	lines := g.Lines()
	plain := func(i int) string { return strings.TrimRight(text([]string{lines[i]}), " \n") }
	if got := plain(0); got != "▸ Accent  #ff007f" {
		t.Errorf("row 0 = %q", got)
	}
	if got := plain(1); got != "  Focus   #00f0ff" {
		t.Errorf("row 1 = %q", got)
	}
	// The selected row is inverted across the whole width; the normal row is
	// not drawn behind its cells.
	if c := g.At(29, 0); c.Bg != "212" {
		t.Errorf("selected row background at the end = %q", c.Bg)
	}
	if c := g.At(29, 1); c.Bg != "" {
		t.Errorf("normal row has a background %q", c.Bg)
	}
}

func TestRowsKeepWhitespaceAndEmptyCells(t *testing.T) {
	g := rowsGrid(20, 2, map[string]string{
		"columns": "A:6,B:6,C",
		"rows":    "  x,,z; y\\,w,,",
	})
	lines := g.Lines()
	first := strings.TrimRight(text([]string{lines[0]}), " \n")
	if first != "  x         z" {
		t.Errorf("leading spaces and the empty cell are kept: %q", first)
	}
	second := strings.TrimRight(text([]string{lines[1]}), " \n")
	if second != " y,w" {
		t.Errorf("an escaped comma stays in its cell: %q", second)
	}
}

func TestRowsDrawColourCellsAsBlocks(t *testing.T) {
	g := rowsGrid(20, 2, map[string]string{
		"columns": "Name:6,Swatch:4:colour",
		"rows":    "Tea,#ff0000;Milk,",
		"styles":  "normal,dim",
	})
	if c := g.At(6, 0); c.Bg != "#ff0000" || c.Ch != ' ' {
		t.Errorf("colour cell = %+v", c)
	}
	if c := g.At(9, 0); c.Bg != "#ff0000" {
		t.Errorf("the block is as wide as the column: %+v", c)
	}
	if c := g.At(10, 0); c.Bg != "" {
		t.Errorf("the block stops at its width: %+v", c)
	}
	if c := g.At(6, 1); c.Ch != '·' {
		t.Errorf("an empty colour shows a dot: %+v", c)
	}
}

func TestRowsSurviveAnySizeAndInput(t *testing.T) {
	cat := standard.Default()
	for _, props := range []map[string]string{
		nil,
		{"columns": "", "rows": "a,b", "styles": "selected"},
		{"columns": ":::,::colour,Name:-3", "rows": ";;;,,,;\\", "styles": "nonsense,selected,dim,accent,,"},
		{"columns": "A:999", "rows": strings.Repeat("x,", 500)},
	} {
		for _, size := range []design.Rect{{W: 0, H: 0}, {W: 1, H: 1}, {W: 5, H: 2}, {W: 80, H: 40}} {
			_ = RenderNode(design.Node{Component: "lipgloss.rows", Rect: size, Props: props}, cat)
		}
	}
}

// A fill column takes what the others leave and an auto column is as wide as
// its own text, so the tail of a row is flush right whatever its width.
func TestRowsFillAndAutoColumnsKeepTheTailAtTheRightEdge(t *testing.T) {
	g := rowsGrid(24, 2, map[string]string{
		"columns": "Mark:1,Sp:1,Title:fill:ellipsis,Gap:1,Count:auto,Sp:1,Glyph:1",
		"rows":    "▸, ,90s UK Dance Hits and more, ,3/6, ,⠋;✓, ,Short, ,listing, ,⟲",
		"styles":  "normal,normal",
	})
	lines := g.Lines()
	want := []string{"▸ 90s UK Dance Hi… 3/6 ⠋", "✓ Short        listing ⟲"}
	for i := range want {
		if got := strings.TrimRight(text([]string{lines[i]}), "\n "); got != want[i] {
			t.Errorf("row %d = %q, want %q", i, got, want[i])
		}
	}
}

// The widths of the columns of a list whose text decides them: fit columns
// from their longest cell within bounds, one that hides when it is empty, and
// the fill column with what is left.
func TestRowsFitColumnsFollowTheirContent(t *testing.T) {
	columns := rowsColumns("Review:9,Artist:fit=6..30%,Song:fit=8..50%,Video:fill:min=4,Replaced:fit=0..25%:floor=8:hide-empty,Address:31")
	rows := [][]string{
		{"", strings.Repeat("a", 12), strings.Repeat("s", 30), strings.Repeat("v", 40), "", ""},
		{"", "b", "c", "d", "", ""},
	}
	plan := newRowsPlan(columns, rows, 100, 2)
	want := []int{9, 12, 26, 14, 0, 31}
	for i, w := range plan.row(rows[0]) {
		if w != want[i] {
			t.Errorf("without replaced titles: column %d is %d wide, want %d (%v)", i, w, want[i], plan.row(rows[0]))
			break
		}
	}
	rows[1][4] = strings.Repeat("r", 20)
	plan = newRowsPlan(columns, rows, 100, 2)
	want = []int{9, 12, 25, 4, 12, 31}
	for i, w := range plan.row(rows[0]) {
		if w != want[i] {
			t.Errorf("with a replaced title: column %d is %d wide, want %d (%v)", i, w, want[i], plan.row(rows[0]))
			break
		}
	}
}

// A title with double-width characters is cut and padded by cells, so the
// tail after it stays in the last cells of the line.
func TestRowsCountCellsNotRunesInAFillColumn(t *testing.T) {
	g := rowsGrid(40, 1, map[string]string{
		"columns": "Glyph:1,Sp:1,Idx:2,Sp2:1,Title:fill:ellipsis,Sp3:1,Tag:auto",
		"rows":    "⠋, ,05, ,Stardust - Music 日本語タイトル と もっと, , 34%",
	})
	line := g.Lines()[0]
	if w := lipgloss.Width(line); w != 40 {
		t.Fatalf("the line is %d cells wide, want 40: %q", w, line)
	}
	plain := strings.TrimRight(text([]string{line}), "\n")
	if !strings.HasSuffix(plain, " 34%") || !strings.Contains(plain, "…") {
		t.Errorf("tail or ellipsis lost: %q", plain)
	}
}
