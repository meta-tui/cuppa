package gosource

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// TestCuppaShellExampleIsResponsive is the dogfood check of the responsive
// layout: examples/cuppa-shell.cuppa is Cuppa's own screen (menu bar, left
// bar, canvas, details bar, status bar) built with expressions.
func TestCuppaShellExampleIsResponsive(t *testing.T) {
	doc, err := disk.Load("../../../examples/cuppa-shell.cuppa")
	if err != nil {
		t.Fatal(err)
	}
	cat := standard.Default()
	ed := editor.New(cat, doc)
	rects := func() map[string]design.Rect {
		out := map[string]design.Rect{}
		for _, n := range ed.Document().Nodes {
			out[n.Name] = n.Rect
		}
		return out
	}
	for _, c := range []struct {
		w, h int
		want map[string]design.Rect
	}{
		{80, 24, map[string]design.Rect{
			"Menu bar":             {X: 0, Y: 0, W: 80, H: 1},
			"Tools and components": {X: 0, Y: 1, W: 24, H: 22},
			"Canvas":               {X: 24, Y: 1, W: 22, H: 22},
			"Details":              {X: 46, Y: 1, W: 34, H: 22},
			"Status bar":           {X: 0, Y: 23, W: 80, H: 1},
		}},
		{160, 50, map[string]design.Rect{
			"Menu bar":             {X: 0, Y: 0, W: 160, H: 1},
			"Tools and components": {X: 0, Y: 1, W: 24, H: 48},
			"Canvas":               {X: 24, Y: 1, W: 102, H: 48},
			"Details":              {X: 126, Y: 1, W: 34, H: 48},
			"Status bar":           {X: 0, Y: 49, W: 160, H: 1},
		}},
	} {
		if err := ed.SetCanvasSize(c.w, c.h); err != nil {
			t.Fatal(err)
		}
		got := rects()
		for name, want := range c.want {
			if got[name] != want {
				t.Errorf("%dx%d: %s = %+v, want %+v", c.w, c.h, name, got[name], want)
			}
		}
	}
	src := Generate(ed.Document(), cat).Files["layout.go"]
	if strings.Count(src, "Fit:") != 5 || !strings.Contains(src, "responsive = true") {
		t.Errorf("export is not responsive:\n%s", src)
	}
}
