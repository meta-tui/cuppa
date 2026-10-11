package editor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func raw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// An application's values are what the designer draws, and the design keeps
// the values it has.
func TestLiveValuesAreDrawnAndNeverKept(t *testing.T) {
	doc := design.NewDocument("Run", 60, 10)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 20, H: 3},
		Props: map[string]string{"title": "design"}, Bind: map[string]string{"title": "Title"}, ShowIf: "ShowBox"})
	doc.Add(design.Node{Component: "lipgloss.rows", Name: "Rows", Rect: design.Rect{Y: 4, W: 30, H: 4},
		Props: map[string]string{"columns": "Mark:2,Name:10", "rows": "a,b"}, Bind: map[string]string{"rows": "List"}})
	ed := New(standard.Default(), doc)
	if ed.Live() {
		t.Fatal("live before any connection")
	}

	ed.SetLive("Run", map[string]json.RawMessage{
		"Title":   raw(t, "From the app"),
		"ShowBox": raw(t, true),
		"List": raw(t, []map[string]any{
			{"Mark": "▸", "Name": "one, two", "Style": "selected"},
			{"Mark": "", "Name": "three"},
		}),
	})
	got := ed.RenderDocument()
	if !ed.Live() {
		t.Fatal("not live after SetLive")
	}
	box, rows := got.Nodes[0], got.Nodes[1]
	if box.Props["title"] != "From the app" || box.Hidden {
		t.Errorf("box: %+v", box.Props)
	}
	if rows.Props["rows"] != `▸,one\, two;,three` || rows.Props["styles"] != "selected,normal" {
		t.Errorf("rows: %q styles %q", rows.Props["rows"], rows.Props["styles"])
	}
	if ed.Document().Nodes[0].Props["title"] != "design" || ed.Dirty() {
		t.Error("a live value reached the design")
	}

	ed.SetLive("Run", map[string]json.RawMessage{"ShowBox": raw(t, false)})
	if !ed.RenderDocument().Nodes[0].Hidden {
		t.Error("a false yes/no input does not hide the component")
	}

	ed.ClearLive()
	if ed.Live() || strings.Contains(ed.RenderDocument().Nodes[0].Props["title"], "app") {
		t.Error("clearing live left its values")
	}
}

func TestANumberAppPublishesFeedsTheLayout(t *testing.T) {
	doc := design.NewDocument("Run", 60, 10)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 3}, Layout: design.Layout{W: "$Width"}})
	ed := New(standard.Default(), doc)
	ed.SetLive("Run", map[string]json.RawMessage{"Width": raw(t, 33)})
	if w := ed.RenderDocument().Nodes[0].Rect.W; w != 33 {
		t.Fatalf("width %d, want 33", w)
	}
}
