package gosource

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

func TestTheDrawingBecomesRunsOfCellsThatShareAStyle(t *testing.T) {
	l := drawlayer.New()
	l.Paint(
		drawlayer.Cell{X: 2, Y: 1, Ch: 'a', Fg: "212"}, drawlayer.Cell{X: 3, Y: 1, Ch: 'b', Fg: "212"},
		drawlayer.Cell{X: 4, Y: 1, Ch: 'c', Fg: "5"}, // another colour: a new run
		drawlayer.Cell{X: 9, Y: 1, Ch: 'd', Fg: "5"}, // a gap: a new run
		drawlayer.Cell{X: 2, Y: 2, Ch: 'e', Fg: "5"}, // another row
	)
	runs := drawRuns(l.Encode(), design.Rect{X: 10, Y: 20, W: 50, H: 30})
	if len(runs) != 4 {
		t.Fatalf("%d runs, want 4: %+v", len(runs), runs)
	}
	if r := runs[0]; r.Props["text"] != "ab" || r.Rect.X != 12 || r.Rect.Y != 21 || r.Rect.W != 2 || r.Kind != drawRunKind {
		t.Fatalf("first run = %+v", r)
	}
}
