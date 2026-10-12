package format

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func sample() design.Document {
	d := design.NewDocument("Login", 80, 24)
	d.Add(design.Node{Component: "bubbles.list", Name: "List", Rect: design.Rect{X: 1, Y: 2, W: 20, H: 8}, Props: map[string]string{"title": "Tea"}})
	d.Add(design.Node{Component: "huh.input", Name: "Input", Rect: design.Rect{X: 30, Y: 4, W: 24, H: 3}})
	return d
}

func TestRoundTrip(t *testing.T) {
	want := sample()
	data, err := Encode(want)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte(Magic)) {
		t.Fatalf("missing magic: %q", data[:8])
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip changed the document:\n got %+v\nwant %+v", got, want)
	}
}

func TestDecodeRejects(t *testing.T) {
	good, _ := Encode(sample())
	newer := append([]byte(nil), good...)
	newer[len(Magic)+1] = byte(CurrentVersion + 1)
	zero := append([]byte(nil), good...)
	zero[len(Magic)], zero[len(Magic)+1] = 0, 0
	cases := map[string]struct {
		data []byte
		want error
	}{
		"empty":      {nil, ErrNotCuppa},
		"json":       {[]byte(`{"document":{}}`), ErrNotCuppa},
		"no version": {[]byte(Magic), ErrCorrupt},
		"version 0":  {zero, ErrCorrupt},
		"newer":      {newer, ErrTooNew},
		"bad gzip":   {append([]byte(Magic+"\x00\x01"), "junk"...), ErrCorrupt},
		"truncated":  {good[:len(good)/2], ErrCorrupt},
	}
	for name, c := range cases {
		if _, err := Decode(c.data); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", name, err, c.want)
		}
	}
}

func TestDecodeRepairsImpossibleNodes(t *testing.T) {
	doc := design.Document{Width: 0, Height: -3, Nodes: []design.Node{
		{ID: "a", Rect: design.Rect{W: 5, H: 5}},
		{ID: "a", Rect: design.Rect{W: 5, H: 5}},
		{ID: "", Rect: design.Rect{W: 5, H: 5}},
		{ID: "b", Rect: design.Rect{W: 0, H: 5}},
	}}
	data, _ := Encode(doc)
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Width != 1 || got.Height != 1 || len(got.Nodes) != 1 {
		t.Fatalf("not repaired: %+v", got)
	}
}

func TestMigrationsRunInOrder(t *testing.T) {
	add := func(field string) migration {
		return func(b rawEnvelope) (rawEnvelope, error) {
			return rawEnvelope(string(b[:len(b)-1]) + `,"` + field + `":1}`), nil
		}
	}
	table := map[uint16]migration{1: add("a"), 2: add("b")}
	got, err := migrateWith(table, rawEnvelope(`{"x":0}`), 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]int
	if err := json.Unmarshal(got, &m); err != nil || m["a"] != 1 || m["b"] != 1 {
		t.Fatalf("got %s (%v)", got, err)
	}
	if _, err := migrateWith(table, rawEnvelope(`{}`), 1, 4); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("missing step: got %v", err)
	}
}

func FuzzDecode(f *testing.F) {
	good, _ := Encode(sample())
	f.Add(good)
	f.Add([]byte(Magic))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := Decode(data)
		if err != nil {
			return
		}
		again, err := Encode(doc)
		if err != nil {
			t.Fatalf("decoded document does not re-encode: %v", err)
		}
		back, err := Decode(again)
		if err != nil || !reflect.DeepEqual(back, doc) {
			t.Fatalf("re-encode changed the document: %v", err)
		}
	})
}

func TestLayerFlagsSurviveSaveAndLoad(t *testing.T) {
	d := sample()
	d.Nodes[0].Hidden = true
	d.Nodes[1].Locked = true
	data, err := Encode(d)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Nodes[0].Hidden || got.Nodes[0].Locked || got.Nodes[1].Hidden || !got.Nodes[1].Locked {
		t.Fatalf("flags lost: %+v %+v", got.Nodes[0], got.Nodes[1])
	}
}

func TestDocumentOptionsRoundTripAndBadOnesAreRepaired(t *testing.T) {
	doc := design.NewDocument("opts", 60, 20)
	doc.Background = "#102030"
	doc.HideGrid = true
	doc.Effects = design.Effects{Shadow: true, Scanlines: true, Vignette: true}
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Background != "#102030" || !got.HideGrid || got.Effects != doc.Effects {
		t.Fatalf("options lost: %+v", got)
	}

	bad := design.NewDocument("bad", 100000, 100000)
	bad.Background = "chartreuse"
	data, _ = Encode(bad)
	got, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Background != "" || got.Width != design.MaxWidth || got.Height != design.MaxHeight {
		t.Fatalf("a damaged file is repaired: %+v", got)
	}
}

func TestGroupsRoundTripAndDamagedGroupsAreRepaired(t *testing.T) {
	doc := design.NewDocument("g", 80, 24)
	doc.Add(design.Node{Component: design.GroupComponent, Name: "Group 1", Rect: design.Rect{X: 2, Y: 2, W: 30, H: 6}, BaseW: 30, BaseH: 6,
		Children: []design.Node{
			{ID: "a", Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 3}},
			{ID: "a", Component: "lipgloss.box", Name: "Twin", Rect: design.Rect{X: 12, W: 10, H: 3}},
			{ID: "b", Component: "lipgloss.box", Name: "Flat", Rect: design.Rect{X: 12, W: 0, H: 3}},
		}})
	doc.Add(design.Node{Component: design.GroupComponent, Name: "Empty", Rect: design.Rect{W: 5, H: 5}, BaseW: 5, BaseH: 5})
	doc.Add(design.Node{Component: design.GroupComponent, Name: "No base", Rect: design.Rect{W: 5, H: 5},
		Children: []design.Node{{ID: "z", Component: "lipgloss.box", Rect: design.Rect{W: 1, H: 1}}}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Plain", Rect: design.Rect{W: 5, H: 5}, BaseW: 9,
		Children: []design.Node{{ID: "z", Component: "lipgloss.box", Rect: design.Rect{W: 1, H: 1}}}})
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Nodes) != 2 {
		t.Fatalf("the good group and the plain node stay: %+v", got.Nodes)
	}
	g := got.Nodes[0]
	if len(g.Children) != 1 || g.Children[0].Name != "Box" || g.BaseW != 30 {
		t.Fatalf("children are cleaned: %+v", g.Children)
	}
	if p := got.Nodes[1]; p.Children != nil || p.BaseW != 0 {
		t.Fatalf("a plain node cannot hold children: %+v", p)
	}
}

func TestEmbeddedComponentsRoundTripAndBadCopiesAreDropped(t *testing.T) {
	good := design.Composite{ID: "card", Name: "Card", W: 10, H: 3,
		Nodes: []design.Node{{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 10, H: 3}}}}
	bad := design.Composite{ID: "Not Valid", Name: "Bad", W: 10, H: 3}
	doc := design.NewDocument("e", 40, 10)
	doc.Embedded = []design.Embedded{{ID: "tea.card", Composite: good}, {ID: "tea.card", Composite: good}, {ID: "tea.bad", Composite: bad}, {ID: "", Composite: good}}
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Embedded) != 1 || got.Embedded[0].ID != "tea.card" || got.Embedded[0].Composite.W != 10 {
		t.Fatalf("embedded = %+v", got.Embedded)
	}
}

func TestProfileAndLightRoundTripAndABadProfileIsDropped(t *testing.T) {
	doc := design.NewDocument("p", 40, 10)
	doc.Profile, doc.Light = design.Profile16, true
	data, _ := Encode(doc)
	got, err := Decode(data)
	if err != nil || got.Profile != design.Profile16 || !got.Light {
		t.Fatalf("round trip: %v %+v", err, got)
	}
	doc.Profile = "sepia"
	data, _ = Encode(doc)
	if got, _ = Decode(data); got.Profile != "" {
		t.Fatalf("an unknown profile is dropped: %q", got.Profile)
	}
}
func TestTheThemeRoundTripsAndABadThemeColourIsDropped(t *testing.T) {
	doc := design.NewDocument("themed", 40, 10)
	doc.Theme = design.Theme{Text: "#112233", Muted: "245", Border: "not a colour", Secondary: "#aa5500"}
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Own", Rect: design.Rect{W: 5, H: 3}, Props: map[string]string{"color": "212"}})
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme.Text != "#112233" || got.Theme.Muted != "245" || got.Theme.Secondary != "#aa5500" {
		t.Errorf("theme = %+v", got.Theme)
	}
	if got.Theme.Border != "" {
		t.Errorf("a bad colour should be dropped: %q", got.Theme.Border)
	}
	if got.Nodes[0].Props["color"] != "212" {
		t.Errorf("the override is stored on the node: %v", got.Nodes[0].Props)
	}
	// A design with no theme stores no theme at all.
	plain, _ := Encode(design.NewDocument("plain", 10, 5))
	if raw, _ := Decode(plain); !raw.Theme.IsZero() {
		t.Error("no theme in, none out")
	}
}

func TestLayoutExpressionsRoundTrip(t *testing.T) {
	d := design.NewDocument("Shell", 120, 40)
	d.Add(design.Node{Component: "lipgloss.box", Name: "Top bar", Rect: design.Rect{W: 120, H: 3}, Layout: design.Layout{X: "0", Y: "0", W: "100%", H: "3"}})
	data, err := Encode(d)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, d) {
		t.Fatalf("layout lost:\n got %+v\nwant %+v", got, d)
	}
}


func TestPaletteRoundTripAndCleaning(t *testing.T) {
	d := design.NewDocument("Colours", 40, 10)
	d.Theme.Palette = []design.Swatch{{Name: "Accent", Color: "#ff007f"}, {Name: "Dim", Color: "#6b7280"}}
	d.Add(design.Node{Component: "lipgloss.label", Name: "Title", Rect: design.Rect{W: 10, H: 1},
		Props: map[string]string{"color": "@Accent"}})
	data, err := Encode(d)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, d) {
		t.Fatalf("palette lost: got %+v, want %+v", got.Theme, d.Theme)
	}

	d.Theme.Palette = []design.Swatch{{Name: "Ok", Color: "#112233"}, {Name: "Ok", Color: "#445566"}, {Name: "9bad", Color: "#000000"}, {Name: "NoColour", Color: "nope"}}
	data, _ = Encode(d)
	got, _ = Decode(data)
	if len(got.Theme.Palette) != 1 || got.Theme.Palette[0].Name != "Ok" || got.Theme.Palette[0].Color != "#112233" {
		t.Errorf("not cleaned: %+v", got.Theme.Palette)
	}
}

// A file made before screens had inputs and events still opens: the fields that
// no longer exist are ignored.
func TestAFileWithFieldsThatNoLongerExistStillOpens(t *testing.T) {
	body := `{"document":{"name":"Old","width":40,"height":10,"keys":[{"key":"s","event":"Save"}],` +
		`"samples":{"Rows":"3"},"defaults":{"A":"1"},"nodes":[{"id":"n1","component":"lipgloss.label","name":"L",` +
		`"rect":{"X":0,"Y":0,"W":10,"H":1},"props":{"text":"Hi"},"bind":{"text":"Title"},"showIf":"Show",` +
		`"event":"Open","draggable":true,"resizable":true}]}}`
	var out bytes.Buffer
	out.WriteString(Magic)
	_ = binary.Write(&out, binary.BigEndian, CurrentVersion)
	zw := gzip.NewWriter(&out)
	_, _ = zw.Write([]byte(body))
	_ = zw.Close()
	doc, err := Decode(out.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name != "Old" || len(doc.Nodes) != 1 || doc.Nodes[0].Props["text"] != "Hi" {
		t.Fatalf("got %+v", doc)
	}
}
