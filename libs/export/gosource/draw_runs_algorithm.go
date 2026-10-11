package gosource

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// drawRunKind is the generated component a run of drawn cells becomes.
const drawRunKind = "draw.run"

// drawRuns splits the drawing into horizontal runs of cells that sit next to
// each other and share a foreground and background colour. origin is where the
// drawing's rectangle sits on the canvas.
func drawRuns(encoded string, origin design.Rect) []leaf {
	var out []leaf
	var cur *leaf
	var text strings.Builder
	var lastX, lastY int
	flush := func() {
		if cur == nil {
			return
		}
		cur.Props["text"] = text.String()
		cur.Rect.W = len([]rune(cur.Props["text"]))
		out = append(out, *cur)
		cur = nil
		text.Reset()
	}
	for _, c := range drawlayer.Decode(encoded).Cells() {
		same := cur != nil && c.Y == lastY && c.X == lastX+1 && c.Fg == cur.Props["fg"] && c.Bg == cur.Props["bg"]
		if !same {
			flush()
			cur = &leaf{
				Kind:  drawRunKind,
				Name:  "Drawing",
				Rect:  design.Rect{X: origin.X + c.X, Y: origin.Y + c.Y, W: 1, H: 1},
				Props: map[string]string{"fg": c.Fg, "bg": c.Bg},
			}
		}
		text.WriteRune(c.Ch)
		lastX, lastY = c.X, c.Y
	}
	flush()
	return out
}
