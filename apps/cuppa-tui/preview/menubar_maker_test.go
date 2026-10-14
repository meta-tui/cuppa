package preview

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestAMenuBarOpensItsMenuWhenPressedInThePreview(t *testing.T) {
	s, _ := newSession(t, node("cuppa.menubar", "Menu bar", design.Rect{X: 2, Y: 1, W: 40, H: 10}, nil))
	if !Live("cuppa.menubar") {
		t.Fatal("the menu bar should run in the preview")
	}
	before := text(s.Render())
	if !strings.Contains(before, "File") || strings.Contains(before, "Open…") {
		t.Fatalf("a closed bar shows only the menu names:\n%s", before)
	}
	// " File " starts at column 1 of the bar, which is at column 2 of the canvas.
	settle(s, s.Press(2+2, 1))
	if out := text(s.Render()); !strings.Contains(out, "Open…") || !strings.Contains(out, "Ctrl+O") {
		t.Fatalf("pressing File opens its box:\n%s", out)
	}
}
