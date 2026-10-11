package editor

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestColourPropertiesAcceptPaletteNumbersAndHexAndStoreTheCanonicalForm(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 80, 24))
	id, err := ed.Add("lipgloss.box", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"212":     "212",
		" 7 ":     "7",
		"#FF8800": "#ff8800",
		"#f80":    "#ff8800",
		"":        "",
	}
	for in, want := range cases {
		if err := ed.SetProp(id, "color", in); err != nil {
			t.Fatalf("%q rejected: %v", in, err)
		}
		n, _ := ed.Document().Get(id)
		got, set := n.Props["color"]
		if !set {
			got = "212" // the default stays implicit
		}
		if want == "212" && !set {
			continue
		}
		if got != want {
			t.Errorf("%q stored as %q, want %q", in, got, want)
		}
	}
}

func TestColourPropertiesRejectWhatATerminalCannotShow(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 80, 24))
	id, _ := ed.Add("lipgloss.box", 1, 1)
	for _, bad := range []string{"red", "256", "#12", "#gg0000", "rgb(1,2,3)"} {
		err := ed.SetProp(id, "color", bad)
		if err == nil {
			t.Errorf("%q should be rejected", bad)
			continue
		}
		if !strings.Contains(err.Error(), "not a colour") {
			t.Errorf("%q: unhelpful error %v", bad, err)
		}
	}
	if n, _ := ed.Document().Get(id); n.Props["color"] != "" {
		t.Errorf("a rejected colour must leave the property alone, got %q", n.Props["color"])
	}
}
