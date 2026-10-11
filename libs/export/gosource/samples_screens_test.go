package gosource

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// An input that only an expression reads starts from its sample value, and
// keeps its type, so the program's models keep compiling.
func TestASampleIsTheDefaultOfAnInputOnlyAnExpressionReads(t *testing.T) {
	doc := design.NewDocument("Download", 100, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Left", Rect: design.Rect{W: 10, H: 5}, Layout: design.Layout{W: "$LeftWidth"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Notes", Rect: design.Rect{Y: 6, W: 10, H: 5}, ShowIf: "$HasNotes > 0"})
	without := GenerateScreens([]design.Document{doc}, standard.Default(), "screens")
	c := without.Files["download_screen_contract.go"]
	if !strings.Contains(c, "LeftWidth int") || !strings.Contains(c, "LeftWidth: 0,") {
		t.Fatalf("without samples:\n%s", c)
	}
	if !strings.Contains(strings.Join(without.Notes, "\n"), "sample value") {
		t.Errorf("the note should say how to give it a value: %v", without.Notes)
	}
	doc.Samples = map[string]string{"LeftWidth": "38", "HasNotes": "true"}
	with := GenerateScreens([]design.Document{doc}, standard.Default(), "screens")
	c = strings.Join(strings.Fields(with.Files["download_screen_contract.go"]), " ")
	for _, want := range []string{"LeftWidth int", "LeftWidth: 38,", "HasNotes int", "HasNotes: 1,"} {
		if !strings.Contains(c, want) {
			t.Errorf("contract lacks %q:\n%s", want, c)
		}
	}
	if len(with.Notes) != 0 {
		t.Errorf("notes: %v", with.Notes)
	}
}
