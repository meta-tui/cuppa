package component

import (
	"os"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

func libraryDescription(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/colourpicker.component.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTheColourPickersDescriptionIsClean(t *testing.T) {
	d, issues, err := Decode(libraryDescription(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("issues: %v", issues)
	}
	if d.ID() != "community.colourpicker" || d.Go.Import != "github.com/meta-ui-labs/bubble-colourpicker" {
		t.Errorf("id %q import %q", d.ID(), d.Go.Import)
	}
	def := d.Definition()
	if def.DefaultSize.W != 46 || len(def.Props) != 5 || def.Props[2].Kind != "choice" {
		t.Errorf("definition %+v", def)
	}
}

// What the module says about the component is what Cuppa's catalog says: the
// built-in entry is the description, so a tool and the designer agree.
func TestTheBuiltInColourPickerMatchesItsDescription(t *testing.T) {
	d, _, err := Decode(libraryDescription(t))
	if err != nil {
		t.Fatal(err)
	}
	built, ok := standard.Default().Get("lipgloss.colourpicker")
	if !ok {
		t.Fatal("the catalog lacks lipgloss.colourpicker")
	}
	want := d.Definition()
	if built.Import != want.Import {
		t.Errorf("import %q, description says %q", built.Import, want.Import)
	}
	if built.DefaultSize != want.DefaultSize || built.MinSize != want.MinSize {
		t.Errorf("size %+v / %+v, description says %+v / %+v", built.DefaultSize, built.MinSize, want.DefaultSize, want.MinSize)
	}
	if len(built.Props) != len(want.Props) {
		t.Fatalf("%d properties, description says %d", len(built.Props), len(want.Props))
	}
	for i, p := range want.Props {
		b := built.Props[i]
		if b.Key != p.Key || b.Kind != p.Kind || b.Default != p.Default || strings.Join(b.Choices, ",") != strings.Join(p.Choices, ",") || b.Min != p.Min || b.Max != p.Max {
			t.Errorf("property %s: catalog %+v, description %+v", p.Key, b, p)
		}
	}
}

func TestADescriptionThatIsWrongSaysWhy(t *testing.T) {
	cases := []struct {
		name, from, to, issue string
	}{
		{"version", `"cuppa": 1`, `"cuppa": 7`, "cuppa is 7"},
		{"name", `"name": "colourpicker"`, `"name": "Colour Picker"`, "must be lowercase"},
		{"licence", `"license": "MIT"`, `"license": ""`, "license is empty"},
		{"import", `"import": "github.com/meta-ui-labs/bubble-colourpicker"`, `"import": ""`, "go.import is empty"},
		{"bubbletea", `"bubbletea": 2`, `"bubbletea": 3`, "go.bubbletea is 3"},
		{"kind", `"kind": "choice"`, `"kind": "dropdown"`, `kind "dropdown"`},
		{"default", `"default": "RGB"`, `"default": "XYZ"`, `default "XYZ" is not one of the choices`},
		{"size", `"min": { "w": 46, "h": 4 }`, `"min": { "w": 99, "h": 4 }`, "size.min is larger"},
		{"event prop", `"prop": "value"`, `"prop": "nothing"`, `property "nothing"`},
		{"identifier", `"constructor": "New"`, `"constructor": "New Thing"`, "is not a Go identifier"},
		{"unknown field", `"homepage"`, `"homepge"`, "unknown or misplaced field"},
	}
	good := string(libraryDescription(t))
	for _, c := range cases {
		if !strings.Contains(good, c.from) {
			t.Fatalf("%s: the description has no %s", c.name, c.from)
		}
		_, issues, err := Decode([]byte(strings.Replace(good, c.from, c.to, 1)))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !strings.Contains(strings.Join(issues, "\n"), c.issue) {
			t.Errorf("%s: issues %v lack %q", c.name, issues, c.issue)
		}
	}
	if _, _, err := Decode([]byte("not json")); err == nil {
		t.Error("a file that is not JSON must be an error")
	}
}
