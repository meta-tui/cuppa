package menubar

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
			Kind    string `json:"kind"`
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
	options := map[string]bool{"WithMenusJSON": true, "WithTitle": true, "WithOpen": true,
		"WithAccent": true, "WithMuted": true, "WithRule": true}
	setters := map[string]bool{"SetMenusJSON": true, "SetTitle": true, "SetOpen": true,
		"SetAccent": true, "SetMuted": true, "SetRule": true}
	getters := map[string]bool{"MenusJSON": true, "Open": true}
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
		if p.Kind == "menu" {
			if _, err := ParseMenus(p.Default); err != nil {
				t.Errorf("prop %s: the default is not menus: %v", p.Key, err)
			}
		}
	}
	for _, e := range d.Events {
		if e.Go.Message != "SelectedMsg" {
			t.Errorf("event message %q does not exist", e.Go.Message)
		}
	}
	// A component built only from the defaults of the description draws.
	var menus string
	for _, p := range d.Props {
		if p.Key == "menus" {
			menus = p.Default
		}
	}
	m := New(WithMenusJSON(menus), WithSize(48, 10))
	if len(m.Menus()) != 3 || len(strings.Split(m.View(), "\n")) != 10 || !strings.Contains(ansi.Strip(m.View()), "Edit") {
		t.Errorf("the defaults of the description do not build: %d menus", len(m.Menus()))
	}
}
