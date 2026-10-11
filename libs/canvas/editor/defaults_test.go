package editor

import (
	"slices"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// An app default belongs to the page it was set on, can be undone, and does not
// change what the designer shows.
func TestAnAppDefaultBelongsToItsPageAndCanBeUndone(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 100, 30))
	id, _ := ed.Add("lipgloss.box", 0, 0)
	if err := ed.SetLayout(id, AxisW, "$FooterLines"); err != nil {
		t.Fatal(err)
	}
	if got := ed.InputNames(); !slices.Equal(got, []string{"FooterLines"}) {
		t.Fatalf("inputs = %v", got)
	}
	if err := ed.SetDefault("FooterLines", "0"); err != nil {
		t.Fatal(err)
	}
	if got := ed.Defaults()["FooterLines"]; got != "0" {
		t.Fatalf("default = %q", got)
	}
	if err := ed.AddPage("Second"); err != nil {
		t.Fatal(err)
	}
	if len(ed.Defaults()) != 0 {
		t.Fatal("the new page has the first page's defaults")
	}
	ed.SwitchPage(0)
	if got := ed.Defaults()["FooterLines"]; got != "0" {
		t.Fatalf("back on the first page: %q", got)
	}
	if err := ed.SetDefault("FooterLines", "two\nlines"); err == nil {
		t.Fatal("a default of two lines was accepted")
	}
}

func TestAnAppDefaultIsOneUndoStep(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 100, 30))
	if err := ed.SetDefault("ShowNote", "false"); err != nil {
		t.Fatal(err)
	}
	ed.Undo()
	if len(ed.Defaults()) != 0 {
		t.Fatalf("undo left %v", ed.Defaults())
	}
}
