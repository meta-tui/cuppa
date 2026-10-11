package editor

import (
	"slices"
	"testing"

	"github.com/meta-tui/cuppa/libs/catalog/standard"
	"github.com/meta-tui/cuppa/libs/document/design"
)

func TestASampleValueGivesAnExpressionItsInput(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 100, 30))
	id, _ := ed.Add("lipgloss.box", 0, 0)
	if err := ed.SetLayout(id, AxisW, "$LeftWidth"); err != nil {
		t.Fatal(err)
	}
	if err := ed.SetLayout(id, AxisH, "10"); err != nil {
		t.Fatal(err)
	}
	if got := ed.ReadInputs(); !slices.Equal(got, []string{"LeftWidth"}) {
		t.Fatalf("read inputs = %v", got)
	}
	before, _ := ed.Document().Get(id)
	if err := ed.SetSample("LeftWidth", "38"); err != nil {
		t.Fatal(err)
	}
	after, _ := ed.Document().Get(id)
	if after.Rect.W != 38 || before.Rect.W == 38 {
		t.Fatalf("width %d -> %d, want 38", before.Rect.W, after.Rect.W)
	}
	if ed.Samples()["LeftWidth"] != "38" {
		t.Errorf("samples = %v", ed.Samples())
	}
	ed.Undo()
	if undone, _ := ed.Document().Get(id); undone.Rect.W == 38 {
		t.Error("undo leaves the sample")
	}
	if err := ed.SetSample("LeftWidth", "wide"); err == nil {
		t.Error("a sample is a number")
	}
	if err := ed.SetSample("not an input!", "1"); err == nil {
		t.Error("a name must be an input name")
	}
}

func TestABoundInputKeepsItsPropertysValueOverASample(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 100, 30))
	box, _ := ed.Add("lipgloss.box", 0, 0)
	if err := ed.SetLayout(box, AxisW, "$Width"); err != nil {
		t.Fatal(err)
	}
	label, _ := ed.Add("bubbles.progress", 0, 5)
	if err := ed.SetBinding(label, "percent", "Width"); err != nil {
		t.Fatal(err)
	}
	if len(ed.ReadInputs()) != 0 {
		t.Fatalf("a bound input is not a sample candidate: %v", ed.ReadInputs())
	}
	if err := ed.SetSample("Width", "70"); err != nil {
		t.Fatal(err)
	}
	n, _ := ed.Document().Get(box)
	if n.Rect.W == 70 {
		t.Error("the bound property's value wins")
	}
}
