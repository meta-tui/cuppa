package editor

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestMenuPropertyStoresTheCanonicalFormAndRefusesWhatIsNotMenus(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 80, 24))
	id, err := ed.Add("cuppa.menubar", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	loose := ` [ {"name" : "File", "entries": [ {"name":"New", "shortcut":"Ctrl+N"}, {"separator": true} ]}, {"name":"Help"} ] `
	if err := ed.SetProp(id, "menus", loose); err != nil {
		t.Fatalf("menus rejected: %v", err)
	}
	n, _ := ed.Document().Get(id)
	want := `[{"name":"File","entries":[{"name":"New","shortcut":"Ctrl+N"},{"separator":true}]},{"name":"Help","entries":[]}]`
	if n.Props["menus"] != want {
		t.Errorf("stored %s\nwant   %s", n.Props["menus"], want)
	}
	for _, bad := range []string{`[{"name":"","entries":[]}]`, `[{"name":"A","entries":[{"shortcut":"x"}]}]`, `plain text`} {
		err := ed.SetProp(id, "menus", bad)
		if err == nil || !strings.Contains(err.Error(), "Menus") {
			t.Errorf("%s was accepted: %v", bad, err)
		}
	}
	if n, _ = ed.Document().Get(id); n.Props["menus"] != want {
		t.Error("a refused value changed the stored one")
	}
}
