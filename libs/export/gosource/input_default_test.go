package gosource

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// An app default replaces the design's value in Default<Screen>Props and only
// there: an input without one keeps the value the design shows.
func TestAnAppDefaultIsWhatDefaultStartsFrom(t *testing.T) {
	doc := design.NewDocument("Download", 100, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Left", Rect: design.Rect{W: 10, H: 5}, Layout: design.Layout{W: "$FooterLines"}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Note", Rect: design.Rect{Y: 6, W: 10, H: 3}, ShowIf: "ShowNote"})
	doc.Samples = map[string]string{"FooterLines": "2"}

	view := func(d design.Document) string {
		t.Helper()
		p := GenerateScreens([]design.Document{d}, standard.Default(), "screens")
		return strings.Join(strings.Fields(p.Files["download_screen_contract.go"]), " ")
	}
	if src := view(doc); !strings.Contains(src, "FooterLines: 2,") || !strings.Contains(src, "ShowNote: true,") {
		t.Fatalf("without defaults the design's values:\n%s", src)
	}

	doc.Defaults = map[string]string{"FooterLines": "(empty)", "ShowNote": "false"}
	if src := view(doc); !strings.Contains(src, "FooterLines: 0,") || !strings.Contains(src, "ShowNote: false,") {
		t.Errorf("the app defaults:\n%s", src)
	}

	doc.Defaults = map[string]string{"FooterLines": "4", "ShowNote": "maybe"}
	p := GenerateScreens([]design.Document{doc}, standard.Default(), "screens")
	src := strings.Join(strings.Fields(p.Files["download_screen_contract.go"]), " ")
	if !strings.Contains(src, "FooterLines: 4,") || !strings.Contains(src, "ShowNote: true,") {
		t.Errorf("a number, and a default that is not a bool keeps the design's value:\n%s", src)
	}
	if !strings.Contains(strings.Join(p.Notes, "\n"), `"maybe"`) {
		t.Errorf("a default that does not fit is noted: %v", p.Notes)
	}
}
