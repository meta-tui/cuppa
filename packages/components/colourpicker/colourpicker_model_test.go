package colourpicker

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func key(s string) tea.KeyPressMsg {
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "shift+right":
		return tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}
	}
	panic("unknown key " + s)
}

func TestValueIsNormalisedAndBadOnesAreIgnored(t *testing.T) {
	m := New(WithValue("#F0A"))
	if m.Value() != "#ff00aa" {
		t.Errorf("short hex = %q", m.Value())
	}
	if m = m.SetValue("212"); m.Value() != "212" {
		t.Errorf("palette = %q", m.Value())
	}
	for _, bad := range []string{"300", "#12", "red", "#gggggg", "-1"} {
		if got := m.SetValue(bad).Value(); got != "212" {
			t.Errorf("SetValue(%q) changed the value to %q", bad, got)
		}
	}
	if m.SetValue("").Value() != "" {
		t.Error("an empty value means no colour")
	}
}

func TestHexIsTheColourWhateverTheValueIs(t *testing.T) {
	cases := []struct{ value, want string }{
		{"", ""},
		{"#ff007f", "#ff007f"},
		{"#f0a", "#ff00aa"},
		{"9", "#ff0000"},
		{"16", "#000000"},
		{"21", "#0000ff"},
		{"232", "#080808"},
		{"255", "#eeeeee"},
	}
	for _, c := range cases {
		if got := New(WithValue(c.value)).Hex(); got != c.want {
			t.Errorf("Hex of %q = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestTabsAreChosenByNameAndKeepTheirOrder(t *testing.T) {
	m := New(WithTabs("hsl", "16", "bogus", "16"), WithTab("16"))
	if got := strings.Join(m.Tabs(), ","); got != "HSL,16" {
		t.Errorf("tabs = %q", got)
	}
	if m.Tab() != Tab16 {
		t.Errorf("tab = %q", m.Tab())
	}
	if m = m.SetTab("256"); m.Tab() != Tab16 {
		t.Error("a tab that is not shown cannot be chosen")
	}
	m, _ = m.Update(key("tab"))
	if m.Tab() != TabHSL {
		t.Errorf("tab key = %q", m.Tab())
	}
	m, _ = m.Update(key("shift+tab"))
	if m.Tab() != Tab16 {
		t.Errorf("shift+tab = %q", m.Tab())
	}
}

func TestSlidersFollowTheKeys(t *testing.T) {
	m := New(WithValue("#102030"), WithTab(TabRGB))
	m, cmd := m.Update(key("right"))
	if m.Value() != "#112030" {
		t.Errorf("right = %q", m.Value())
	}
	if cmd == nil {
		t.Fatal("a change returns a command")
	}
	if msg, ok := cmd().(ChangedMsg); !ok || msg.Value != "#112030" {
		t.Errorf("message = %#v", cmd())
	}
	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("shift+right"))
	if m.Value() != "#112a30" || m.Slide() != 1 {
		t.Errorf("down, shift+right = %q on slider %d", m.Value(), m.Slide())
	}
	if _, cmd := m.Update(key("up")); cmd != nil {
		t.Error("moving between sliders is not a change")
	}
}

func TestHSLKeepsTheHueOfAGrey(t *testing.T) {
	m := New(WithValue("#808080"), WithTab(TabHSL))
	m, _ = m.Update(key("shift+right")) // hue 10 of a grey: no visible change
	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("shift+right")) // saturation 10: the hue must still be 10
	if h := m.hslNow().h; h != 10 {
		t.Errorf("hue after raising the saturation = %.1f, want 10", h)
	}
}

func TestNoColourStartsTheHSLSlidersFromThePink(t *testing.T) {
	plain := ansi.Strip(New(WithTab(TabHSL)).View())
	if !strings.Contains(plain, "320°") {
		t.Errorf("the hue of the starting pink is not shown:\n%s", plain)
	}
}

func TestSwatchesMoveByTheirLayout(t *testing.T) {
	m := New(WithValue("0"), WithTab(Tab16))
	steps := []struct {
		key, want string
	}{{"right", "1"}, {"right", "2"}, {"down", "10"}, {"left", "9"}, {"up", "1"}, {"up", "1"}}
	for _, s := range steps {
		m, _ = m.Update(key(s.key))
		if m.Value() != s.want {
			t.Fatalf("after %s: %q, want %q", s.key, m.Value(), s.want)
		}
	}
	// A hex colour is not a swatch: the first arrow picks the first swatch.
	m = New(WithValue("#123456"), WithTab(Tab16))
	if m, _ = m.Update(key("right")); m.Value() != "0" {
		t.Errorf("from a hex colour = %q", m.Value())
	}
}

func TestThe256TabHasAllTheColours(t *testing.T) {
	m := New(WithTab(Tab256))
	_, hits, _ := m.layout()
	seen := map[int]bool{}
	for _, h := range hits {
		if h.kind == pickerHitSwatch {
			seen[h.arg] = true
		}
	}
	if len(seen) != 256 {
		t.Errorf("%d swatches, want 256", len(seen))
	}
	if h := m.Height(); h != 23 {
		t.Errorf("the 256 tab is %d lines, want 23", h)
	}
	if h := New(WithTab(TabRGB)).Height(); h != 10 {
		t.Errorf("the RGB tab is %d lines, want 10", h)
	}
}

func TestMouseClicksAndDrags(t *testing.T) {
	m := New(WithValue("#000000"), WithTab(TabRGB), WithOrigin(10, 5))
	// The red bar starts at column 8 of the picker, in its row 2.
	m, _ = m.Update(tea.MouseClickMsg{X: 10 + 8 + 27, Y: 5 + 2, Button: tea.MouseLeft})
	if m.Value() != "#ff0000" {
		t.Errorf("click at the end of the bar = %q", m.Value())
	}
	m, _ = m.Update(tea.MouseMotionMsg{X: 10 + 8, Y: 5 + 2, Button: tea.MouseLeft})
	if m.Value() != "#000000" {
		t.Errorf("drag to the start = %q", m.Value())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 10 + 4, Y: 5 + 4, Button: tea.MouseLeft}) // [-] of green: stays at 0
	m, _ = m.Update(tea.MouseClickMsg{X: 10 + 8 + 28 + 1, Y: 5 + 4, Button: tea.MouseLeft})
	if m.Value() != "#000100" {
		t.Errorf("[+] of green = %q", m.Value())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 10 + 1, Y: 5, Button: tea.MouseLeft}) // the first tab: 16
	if m.Tab() != Tab16 {
		t.Errorf("click on the first tab = %q", m.Tab())
	}
	m, _ = m.Update(tea.MouseClickMsg{X: 10 + 1, Y: 5 + 2, Button: tea.MouseLeft}) // first swatch
	if m.Value() != "0" {
		t.Errorf("click on a swatch = %q", m.Value())
	}
	// Clicks outside the picker, or with another button, change nothing.
	m2, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if m2.Value() != m.Value() || cmd != nil {
		t.Error("a click elsewhere changed the picker")
	}
	if m3, _ := m.Update(tea.MouseClickMsg{X: 10 + 1 + 6, Y: 5 + 2, Button: tea.MouseRight}); m3.Value() != m.Value() {
		t.Error("the right button changed the picker")
	}
}

func TestViewIsAsWideAsItSaysAndShowsTheValue(t *testing.T) {
	for _, tab := range []string{Tab16, Tab256, TabRGB, TabHSL} {
		m := New(WithValue("#ff007f"), WithTab(tab))
		lines := strings.Split(m.View(), "\n")
		if len(lines) != m.Height() {
			t.Errorf("%s: %d lines, Height says %d", tab, len(lines), m.Height())
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w != m.Width() {
				t.Errorf("%s line %d is %d cells wide, want %d", tab, i, w, m.Width())
			}
		}
		if plain := ansi.Strip(m.View()); !strings.Contains(plain, "Value #ff007f") {
			t.Errorf("%s: the value is not shown:\n%s", tab, plain)
		}
	}
	if plain := ansi.Strip(New().View()); !strings.Contains(plain, "(none)") || !strings.Contains(plain, "Value none") {
		t.Errorf("no colour:\n%s", plain)
	}
}

// The description file for tools must name things that exist.
func TestComponentDescriptionMatchesTheCode(t *testing.T) {
	data, err := os.ReadFile("cuppa.component.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Go    map[string]any `json:"go"`
		Props []struct {
			Key     string   `json:"key"`
			Default string   `json:"default"`
			Choices []string `json:"choices"`
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
	options := map[string]bool{"WithValue": true, "WithTabs": true, "WithTab": true, "WithAccent": true, "WithOrigin": true}
	setters := map[string]bool{"SetValue": true, "SetTab": true, "SetSlide": true, "SetAccent": true, "SetOrigin": true}
	getters := map[string]bool{"Value": true, "Tab": true, "Slide": true, "Tabs": true, "Hex": true}
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
	// The defaults in the description are what New would give.
	m := New()
	for _, p := range d.Props {
		switch p.Key {
		case "tab":
			if p.Default != m.Tab() {
				t.Errorf("default tab %q, New gives %q", p.Default, m.Tab())
			}
		case "tabs":
			if p.Default != strings.Join(m.Tabs(), ",") {
				t.Errorf("default tabs %q, New gives %q", p.Default, strings.Join(m.Tabs(), ","))
			}
		}
	}
}
