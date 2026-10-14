package scene

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func paintedDivider(t *testing.T, w, h int, props Props) []string {
	t.Helper()
	g := grid.New(w, h)
	paintDivider(g, props)
	lines := g.Lines()
	for i := range lines {
		lines[i] = ansi.Strip(lines[i])
	}
	return lines
}

func TestDividerDrawsAHorizontalRuleOnTheMiddleRow(t *testing.T) {
	lines := paintedDivider(t, 10, 3, Props{"orientation": "horizontal", "style": "double"})
	if strings.TrimSpace(lines[0]) != "" || strings.TrimSpace(lines[2]) != "" {
		t.Errorf("only the middle row is drawn: %q", lines)
	}
	if lines[1] != strings.Repeat("═", 10) {
		t.Errorf("middle = %q", lines[1])
	}
}

func TestDividerDrawsAVerticalRuleDownTheMiddleColumn(t *testing.T) {
	lines := paintedDivider(t, 3, 4, Props{"orientation": "vertical", "style": "heavy"})
	for y, l := range lines {
		if l != " ┃ " {
			t.Errorf("row %d = %q", y, l)
		}
	}
}

func TestDividerLabelBreaksTheLineAndGoesWhereItIsAskedFor(t *testing.T) {
	centre := paintedDivider(t, 21, 1, Props{"label": "Settings", "align": "center"})[0]
	if !strings.Contains(centre, " Settings ") || !strings.HasPrefix(centre, "──") || !strings.HasSuffix(centre, "──") {
		t.Errorf("centre = %q", centre)
	}
	left := paintedDivider(t, 21, 1, Props{"label": "Settings", "align": "left"})[0]
	if !strings.HasPrefix(left, "── Settings ") {
		t.Errorf("left = %q", left)
	}
	right := paintedDivider(t, 21, 1, Props{"label": "Settings", "align": "right"})[0]
	if !strings.HasSuffix(right, " Settings ──") {
		t.Errorf("right = %q", right)
	}
	if got := paintedDivider(t, 8, 1, Props{"label": "Settings"})[0]; got != strings.Repeat("─", 8) {
		t.Errorf("a label that does not fit is left out: %q", got)
	}
	if got := paintedDivider(t, 10, 1, Props{"style": "nonsense"})[0]; got != strings.Repeat("─", 10) {
		t.Errorf("an unknown style falls back to single: %q", got)
	}
}
