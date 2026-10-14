package scene

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/bundled"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Every component of the bundled widgets pack draws something recognisable at
// its default size, built only from the components that ship.
func TestWidgetsPackComponentsDrawAtTheirDefaultSize(t *testing.T) {
	packs, problems := bundled.Packs()
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	cat, issues := cupp.Extend(standard.Default(), packs)
	if len(issues) != 0 {
		t.Fatal(issues)
	}
	want := map[string]string{
		"widgets.badge":             "NEW",
		"widgets.key-hint":          "Ctrl+C",
		"widgets.stat-card":         "Cups today",
		"widgets.breadcrumb":        "Home › Teas › Earl Grey",
		"widgets.sidebar-menu":      "Dashboard",
		"widgets.command-palette":   "Type a command",
		"controls.checkbox":         "[x] Steep for three minutes",
		"controls.radio":            "(•) Green tea",
		"controls.toggle":           "[━●] Notifications",
		"controls.accordion-open":   "▾ Details",
		"controls.accordion-closed": "▸ Details",
	}
	for id, text := range want {
		def, ok := cat.Get(id)
		if !ok {
			t.Errorf("%s is missing from the pack", id)
			continue
		}
		doc := design.NewDocument("w", def.DefaultSize.W+2, def.DefaultSize.H+2)
		doc.Add(design.Node{Component: id, Name: def.Name, Rect: design.Rect{X: 1, Y: 1, W: def.DefaultSize.W, H: def.DefaultSize.H}})
		if out := rows(doc, cat); !strings.Contains(out, text) {
			t.Errorf("%s: %q not found in\n%s", id, text, out)
		}
	}
}
