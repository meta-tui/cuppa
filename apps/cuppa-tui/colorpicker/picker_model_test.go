package colorpicker

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/color/scheme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

func open(initial string) *Model {
	m := New("Border color", initial)
	m.Place(120, 40)
	m.Lines()
	return m
}

// press clicks the box-relative cell (x, y).
func press(m *Model, x, y int) {
	m.Handle(pointer.Event{X: m.rect.X + x, Y: m.rect.Y + y, Phase: pointer.Down, Left: true})
	m.Lines()
}

func clickTab(t *testing.T, m *Model, name string) {
	t.Helper()
	for i, n := range tabNames {
		if n == name {
			g := m.regions[i] // the tab bar is registered first
			press(m, g.x0+1, g.y)
			return
		}
	}
	t.Fatalf("no tab %q", name)
}

func TestOpensOnTheStoredColour(t *testing.T) {
	cases := []struct {
		in    string
		tab   tab
		value string
	}{
		{"9", tab16, "9"}, {"212", tab256, "212"}, {"#112233", tabRGB, "#112233"}, {"", tab16, ""},
	}
	for _, c := range cases {
		m := open(c.in)
		if m.tab != c.tab || m.Value() != c.value {
			t.Errorf("%q: tab %v value %q", c.in, m.tab, m.Value())
		}
	}
}

func TestEverySwatchOfTheSixteenIsReachable(t *testing.T) {
	m := open("")
	clickTab(t, m, "16")
	seen := map[string]bool{}
	for _, g := range append([]region(nil), m.regions[len(tabNames):]...) {
		if g.down == nil {
			continue
		}
		press(m, g.x0, g.y)
		seen[m.Value()] = true
	}
	for i := 0; i < 16; i++ {
		if !seen[itoa(i)] {
			t.Errorf("swatch %d cannot be picked", i)
		}
	}
}

func itoa(i int) string { return strconv.Itoa(i) }

func TestEveryColourOfThe256IsReachable(t *testing.T) {
	m := open("")
	clickTab(t, m, "256")
	seen := map[string]bool{}
	for _, g := range append([]region(nil), m.regions[len(tabNames):]...) {
		if g.down == nil || g.y >= bodyTop+m.bodyHeight() {
			continue
		}
		press(m, g.x0, g.y)
		seen[m.Value()] = true
	}
	if len(seen) != 256 {
		t.Fatalf("only %d of 256 colours can be picked", len(seen))
	}
	for _, want := range []string{"0", "15", "16", "212", "231", "255"} {
		if !seen[want] {
			t.Errorf("colour %s missing", want)
		}
	}
}

func TestRGBSlidersSetChannelsByClickAndDrag(t *testing.T) {
	m := open("#000000")
	if m.tab != tabRGB {
		t.Fatal("a hex colour opens on the RGB tab")
	}
	var bars []region
	for _, g := range m.regions {
		if g.drag != nil {
			bars = append(bars, g)
		}
	}
	if len(bars) != 3 {
		t.Fatalf("want 3 bars, got %d", len(bars))
	}
	red := bars[0]
	press(m, red.x0+barWidth-1, red.y) // far right: full red
	if m.Value() != "#ff0000" {
		t.Fatalf("after clicking the end of R: %s", m.Value())
	}
	m.Handle(pointer.Event{X: m.rect.X + red.x0, Y: m.rect.Y + red.y, Phase: pointer.Move, Held: true}) // drag to the start
	if m.Value() != "#000000" {
		t.Fatalf("after dragging to the start: %s", m.Value())
	}
	m.Handle(pointer.Event{Phase: pointer.Up})
	green := bars[1]
	press(m, green.x0+barWidth-1, green.y)
	if m.Value() != "#00ff00" {
		t.Fatalf("G: %s", m.Value())
	}
}

func TestStepButtonsNudgeByOne(t *testing.T) {
	m := open("#102030")
	var plus region
	for _, g := range m.regions {
		if g.down != nil && g.drag == nil && g.y == bodyTop { // the first row's [-] then [+]
			plus = g
		}
	}
	press(m, plus.x0, plus.y)
	if m.Value() != "#112030" {
		t.Fatalf("after [+] on R: %s", m.Value())
	}
}

func TestHSLSlidersChooseAnyColourAndStayConsistent(t *testing.T) {
	m := open("#ff0000")
	clickTab(t, m, "HSL")
	var bars []region
	for _, g := range m.regions {
		if g.drag != nil {
			bars = append(bars, g)
		}
	}
	// Hue a third of the way along is green-ish: 120 degrees is x = 119/359*(barWidth-1) ~ 10.
	press(m, bars[0].x0+10, bars[0].y)
	got, _ := space.Resolve(m.Value())
	if hsl := space.ToHSL(got); hsl.H < 100 || hsl.H > 140 {
		t.Fatalf("hue %v for %s", hsl.H, m.Value())
	}
	if !strings.HasPrefix(m.Value(), "#") {
		t.Fatalf("slider colours are true colour hex: %s", m.Value())
	}
	press(m, bars[2].x0, bars[2].y) // lightness 0: black
	if m.Value() != "#000000" {
		t.Fatalf("lightness 0: %s", m.Value())
	}
}

func TestSwitchingTabsKeepsTheColour(t *testing.T) {
	m := open("212")
	clickTab(t, m, "HSL")
	clickTab(t, m, "RGB")
	if m.Value() != "212" {
		t.Fatalf("looking at other tabs must not change the colour: %s", m.Value())
	}
}

func TestTypingAValueAndEnter(t *testing.T) {
	m := open("")
	for _, r := range "#FF8000" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	out, done := m.Outcome()
	if !done || out.Canceled || out.Value != "#ff8000" || out.Button != "OK" {
		t.Fatalf("got %+v done=%v", out, done)
	}
}

func TestTypingAnInvalidValueExplainsAndStaysOpen(t *testing.T) {
	m := open("")
	for _, r := range "red" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	if _, done := m.Outcome(); done || m.errMsg == "" {
		t.Fatalf("done=%v err=%q", done, m.errMsg)
	}
	m.Lines()
	if !strings.Contains(ansi.Strip(strings.Join(m.Lines(), "\n")), "not a colour") {
		t.Fatal("the error should be shown")
	}
}

func TestButtons(t *testing.T) {
	m := open("#123456")
	var btns []region
	for _, g := range m.regions {
		if g.y == bodyTop+m.bodyHeight()+4 {
			btns = append(btns, g)
		}
	}
	if len(btns) != 3 {
		t.Fatalf("want None, OK and Cancel, got %d", len(btns))
	}
	press(m, btns[1].x0, btns[1].y)
	if out, done := m.Outcome(); !done || out.Value != "#123456" || out.Button != "OK" {
		t.Fatalf("OK: %+v", out)
	}
	m = open("#123456")
	press(m, btns[0].x0, btns[0].y)
	if out, _ := m.Outcome(); out.Value != "" || out.Button != "None" || out.Canceled {
		t.Fatalf("None: %+v", out)
	}
	m = open("#123456")
	press(m, btns[2].x0, btns[2].y)
	if out, _ := m.Outcome(); !out.Canceled {
		t.Fatalf("Cancel: %+v", out)
	}
}

func TestEscCancels(t *testing.T) {
	m := open("5")
	m.Key("", false, false, true)
	if out, done := m.Outcome(); !done || !out.Canceled {
		t.Fatalf("%+v", out)
	}
}

func TestEveryTabRendersExactlyItsRect(t *testing.T) {
	for _, name := range tabNames {
		m := open("212")
		clickTab(t, m, name)
		lines := m.Lines()
		r := m.Rect()
		if len(lines) != r.H {
			t.Errorf("%s: %d lines, rect height %d", name, len(lines), r.H)
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w != r.W {
				t.Errorf("%s line %d: width %d, want %d: %q", name, i, w, r.W, ansi.Strip(l))
			}
		}
	}
}

func TestFirstKeystrokeReplacesTheShownValueAndBackspaceThenEditsNormally(t *testing.T) {
	m := open("212")
	m.Key("#", false, false, false)
	m.Key("a", false, false, false)
	if string(m.text) != "#a" {
		t.Fatalf("typing should replace the pre-filled value, got %q", string(m.text))
	}
	m.Key("", true, false, false)
	if string(m.text) != "#" {
		t.Fatalf("backspace should delete one character, got %q", string(m.text))
	}
	m = open("212")
	m.Key("", true, false, false)
	if len(m.text) != 0 {
		t.Fatalf("backspace on an untouched value clears it, got %q", string(m.text))
	}
}

func TestThemesTabPicksAnyColourOfAScheme(t *testing.T) {
	m := open("")
	clickTab(t, m, "Themes")
	s := scheme.All()[m.scheme]
	seen := map[string]bool{}
	for _, g := range append([]region(nil), m.regions[len(tabNames):]...) {
		if g.down == nil || g.y < bodyTop+4 {
			continue // the step buttons and the letter row
		}
		press(m, g.x0, g.y)
		seen[m.Value()] = true
	}
	for i, c := range s.ANSI {
		if !seen[c.Hex()] && !hasDuplicate(s, i) {
			t.Errorf("%s: colour %d (%s) cannot be picked", s.Name, i, c.Hex())
		}
	}
	if !seen[s.Foreground.Hex()] || !seen[s.Background.Hex()] {
		t.Errorf("%s: text and page colours should be pickable", s.Name)
	}
}

// hasDuplicate reports whether an earlier ANSI colour of the scheme has the same value.
func hasDuplicate(s scheme.Scheme, i int) bool {
	for j := 0; j < i; j++ {
		if s.ANSI[j] == s.ANSI[i] {
			return true
		}
	}
	return false
}

func TestThemesTabStepsAndJumpsBetweenSchemes(t *testing.T) {
	m := open("")
	clickTab(t, m, "Themes")
	start := m.scheme
	m.stepScheme(1)
	if m.scheme != start+1 {
		t.Fatalf("step: %d, want %d", m.scheme, start+1)
	}
	m.stepScheme(-len(scheme.All()))
	if m.scheme != start+1 {
		t.Fatal("stepping a full lap should land on the same scheme")
	}
	m.jumpScheme('N')
	if name := scheme.All()[m.scheme].Name; !strings.HasPrefix(strings.ToUpper(name), "N") {
		t.Fatalf("jump to N landed on %q", name)
	}
	m.jumpScheme('#')
	if name := scheme.All()[m.scheme].Name; name[0] < '0' || name[0] > '9' {
		t.Fatalf("jump to # landed on %q", name)
	}
	before := m.scheme
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1, X: m.rect.X + 5, Y: m.rect.Y + 5})
	if m.scheme != before+1 {
		t.Errorf("wheel: %d, want %d", m.scheme, before+1)
	}
}

func TestPickingASchemeColourStoresAPlainHexColour(t *testing.T) {
	m := open("")
	clickTab(t, m, "Themes")
	m.stepScheme(0)
	s := scheme.All()[m.scheme]
	for _, g := range m.regions[len(tabNames):] {
		if g.down != nil && g.y >= bodyTop+4 {
			press(m, g.x0, g.y)
			break
		}
	}
	if m.Value() != s.ANSI[0].Hex() {
		t.Fatalf("value = %q, want %q", m.Value(), s.ANSI[0].Hex())
	}
	if _, err := space.Normalise(m.Value()); err != nil {
		t.Fatal(err)
	}
}
