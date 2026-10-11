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

// shell is the top bar, a left bar and the main area of an app.
func shell() design.Document {
	doc := design.NewDocument("shell", 120, 40)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Top bar", Rect: design.Rect{W: 120, H: 3},
		Layout: design.Layout{X: "0", Y: "0", W: "100%", H: "3"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Left bar", Rect: design.Rect{Y: 3, W: 30, H: 37},
		Layout: design.Layout{W: "30", H: "100% - 3"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Main", Rect: design.Rect{X: 30, Y: 3, W: 60, H: 37},
		Layout: design.Layout{X: "30", W: "100% - 60", H: "max(10, 100% - 3)"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Fixed", Rect: design.Rect{X: 2, Y: 5, W: 10, H: 1}})
	return doc
}

func TestLayoutExpressionsBecomeFitFunctions(t *testing.T) {
	src := Generate(shell(), standard.Default()).Files["layout.go"]
	for _, want := range []string{
		"responsive = true",
		"Fit: func(w, h int) (x, y, cw, ch int)",
		"cells(((float64(w) * 100.0 / 100.0) - 60.0))",
		"cells(max(10.0, ((float64(h) * 100.0 / 100.0) - 3.0)))",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("layout.go lacks %q:\n%s", want, src)
		}
	}
	if strings.Count(src, "Fit:") != 3 {
		t.Errorf("the fixed label must not have a Fit:\n%s", src)
	}
}

func TestAFixedDesignIsNotResponsive(t *testing.T) {
	doc := design.NewDocument("fixed", 80, 24)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 20, H: 5}})
	src := Generate(doc, standard.Default()).Files["layout.go"]
	if !strings.Contains(src, "responsive = false") || strings.Contains(src, "Fit:") {
		t.Errorf("a fixed design got layout code:\n%s", src)
	}
}

// TestTheGeneratedProgramFollowsTheWindow builds the program and checks the
// rectangles it computes at two window sizes.
func TestTheGeneratedProgramFollowsTheWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(shell(), standard.Default())); err != nil {
		t.Fatal(err)
	}
	check := `package main

import "testing"

func TestRectanglesFollowTheWindow(t *testing.T) {
	m := newModel()
	for _, c := range []struct {
		w, h int
		want [4][4]int // top bar, left bar, main, fixed label: x, y, w, h
	}{
		{80, 24, [4][4]int{{0, 0, 80, 3}, {0, 3, 30, 21}, {30, 3, 20, 21}, {2, 5, 10, 1}}},
		{200, 60, [4][4]int{{0, 0, 200, 3}, {0, 3, 30, 57}, {30, 3, 140, 57}, {2, 5, 10, 1}}},
	} {
		m.resize(c.w, c.h)
		for i, want := range c.want {
			got := [4]int{layout[i].X, layout[i].Y, layout[i].W, layout[i].H}
			if got != want {
				t.Errorf("window %dx%d, %s = %v, want %v", c.w, c.h, layout[i].Name, got, want)
			}
		}
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "resize_test.go"), []byte(check), 0o644); err != nil {
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

// TestTheGeneratedProgramLetsTheUserDragAndResize builds the program and
// drives the mouse handling of a draggable, resizable component that also
// follows the window.
func TestTheGeneratedProgramLetsTheUserDragAndResize(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	doc := design.NewDocument("movable", 80, 24)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Panel", Rect: design.Rect{X: 10, Y: 5, W: 20, H: 6},
		Draggable: true, Resizable: true, Layout: design.Layout{W: "50%"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Still", Rect: design.Rect{X: 50, Y: 15, W: 20, H: 6}})
	src := Generate(doc, standard.Default()).Files["layout.go"]
	if !strings.Contains(src, "Drag: true, Resize: true, MinW: 3, MinH: 3, Fit:") {
		t.Fatalf("flags missing:\n%s", src)
	}
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(doc, standard.Default())); err != nil {
		t.Fatal(err)
	}
	check := `package main

import "testing"

func rectOf(i int) [4]int { return [4]int{layout[i].X, layout[i].Y, layout[i].W, layout[i].H} }

func TestDragAndResize(t *testing.T) {
	m := newModel()
	// Press inside the panel, move 5 right and 2 down.
	m.take(15, 7)
	m.drag(20, 9)
	m.grab = nil
	if got := rectOf(0); got != [4]int{15, 7, 20, 6} {
		t.Fatalf("moved = %v", got)
	}
	// Grab its bottom-right corner and make it 30 wide and 8 tall.
	m.take(15+20-1, 7+6-1)
	m.drag(15+30-1, 7+8-1)
	m.grab = nil
	if got := rectOf(0); got != [4]int{15, 7, 30, 8} {
		t.Fatalf("resized = %v", got)
	}
	// A bigger window moves what Fit gives, and the person's change is kept.
	m.resize(160, 50)
	if got := rectOf(0); got != [4]int{15, 7, 90, 8} {
		t.Fatalf("after the window grew = %v", got)
	}
	// It cannot be dragged off the window, nor made smaller than a box.
	m.take(16, 8)
	m.drag(500, 500)
	m.grab = nil
	if got := rectOf(0); got[0]+got[2] != 160 || got[1]+got[3] != 50 {
		t.Fatalf("left the window: %v", got)
	}
	// A component that is not draggable ignores the mouse.
	before := rectOf(1)
	m.take(layout[1].X+1, layout[1].Y+1)
	if m.grab != nil {
		t.Fatal("grabbed a fixed component")
	}
	if rectOf(1) != before {
		t.Fatal("moved")
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "drag_test.go"), []byte(check), 0o644); err != nil {
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
