package panels

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// The description file for tools must name things that exist.
func TestComponentDescriptionMatchesTheCode(t *testing.T) {
	data, err := os.ReadFile("cuppa.component.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Props []struct {
			Key     string `json:"key"`
			Default string `json:"default"`
			Go      struct {
				Option string `json:"option"`
				Setter string `json:"setter"`
				Getter string `json:"getter"`
			} `json:"go"`
		} `json:"props"`
		Events []struct {
			Go struct {
				Message string `json:"message"`
			} `json:"go"`
		} `json:"events"`
	}
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	options := map[string]bool{"WithModeName": true, "WithTitles": true, "WithActive": true, "WithOpenText": true,
		"WithExclusive": true, "WithAccent": true, "WithMuted": true, "WithRule": true}
	setters := map[string]bool{"SetActive": true, "SetExclusive": true, "SetAccent": true, "SetMuted": true, "SetRule": true}
	getters := map[string]bool{"Active": true, "String": true, "Exclusive": true}
	for _, p := range d.Props {
		if p.Go.Option != "" && !options[p.Go.Option] {
			t.Errorf("prop %s: no option %s", p.Key, p.Go.Option)
		}
		if p.Go.Setter != "" && !setters[p.Go.Setter] {
			t.Errorf("prop %s: no setter %s", p.Key, p.Go.Setter)
		}
		if p.Go.Getter != "" && !getters[p.Go.Getter] {
			t.Errorf("prop %s: no getter %s", p.Key, p.Go.Getter)
		}
	}
	for _, e := range d.Events {
		if e.Go.Message != "ChangedMsg" {
			t.Errorf("event message %q does not exist", e.Go.Message)
		}
	}
	// A component built only from the defaults of the description draws.
	m := New(WithTitles("Details", "Properties", "Layers"), WithOpenText("0,1"), WithSize(40, 16))
	if len(m.OpenSections()) != 2 || len(strings.Split(m.View(), "\n")) != 16 {
		t.Errorf("the defaults of the description do not build: open %v", m.OpenSections())
	}
}

func TestADesignBuildsItFromScalarsOnly(t *testing.T) {
	m := New(WithModeName("tabs"), WithTitles("One", " Two ", ""), WithActive(1), WithSize(30, 6))
	if m.Mode() != Tabs || len(m.Pages()) != 2 || m.Active() != 1 {
		t.Fatalf("mode %v pages %d active %d", m.Mode(), len(m.Pages()), m.Active())
	}
	m.SetBody(1, text("hello"))
	if !strings.Contains(ansi.Strip(m.View()), "hello") {
		t.Errorf("body not drawn:\n%s", ansi.Strip(m.View()))
	}
}
