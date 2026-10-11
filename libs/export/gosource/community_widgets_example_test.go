package gosource

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const communityWidgetsExample = "../../../examples/community-widgets.cuppa"

// communityWidgets builds, through the editor, a design that shows each
// component of the third research round (#181) once.
func communityWidgets(t *testing.T) design.Document {
	t.Helper()
	ed := editor.New(standard.Default(), design.NewDocument("Community widgets", 100, 35))
	place := func(comp, name string, x, y, w, h int, props map[string]string) {
		t.Helper()
		id, err := ed.Add(comp, x, y)
		if err != nil {
			t.Fatal(err)
		}
		ed.SetRect(id, design.Rect{X: x, Y: y, W: w, H: h}, false)
		ed.Rename(id, name)
		for k, v := range props {
			if err := ed.SetProp(id, k, v); err != nil {
				t.Fatal(err)
			}
		}
	}
	place("community.dropdown", "Dropdown", 1, 1, 26, 7, nil)
	place("community.datatree", "Data tree", 1, 10, 34, 9, nil)
	place("community.promptinput", "Name prompt", 38, 1, 36, 2, map[string]string{"prompt": "Tea:", "value": "Sencha"})
	place("community.promptinput", "Password prompt", 38, 4, 36, 2, map[string]string{"prompt": "Key:", "value": "secret", "hidden": "true", "error": "too short"})
	place("community.promptselect", "Tea select", 38, 8, 30, 6, nil)
	place("community.pdfview", "PDF viewer", 1, 20, 44, 14, nil)
	place("ntcharts.chart3d", "3D chart", 50, 20, 44, 14, nil)
	return ed.Document()
}

// TestCommunityWidgetsExample keeps examples/community-widgets.cuppa in step
// with the design above (set CUPPA_WRITE_EXAMPLES=1 to write it) and checks
// that it exports as screens without notes.
func TestCommunityWidgetsExample(t *testing.T) {
	want := communityWidgets(t)
	path := filepath.FromSlash(communityWidgetsExample)
	if os.Getenv("CUPPA_WRITE_EXAMPLES") == "1" {
		if _, err := disk.Save(path, want); err != nil {
			t.Fatal(err)
		}
	}
	got, err := disk.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("examples/community-widgets.cuppa is out of date; run with CUPPA_WRITE_EXAMPLES=1")
	}
	if p := Generate(got, standard.Default()); len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
}
