package shell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func paletteText(m *Model) string {
	var rows []string
	for _, l := range m.pal.Lines() {
		rows = append(rows, ansi.Strip(l))
	}
	return strings.Join(rows, "\n")
}

func TestSwitchingAPackOffRemovesItFromThePaletteOnly(t *testing.T) {
	m := newShell(t)
	m.pal.SetSize(m.layout.palette.W, 80)
	if !strings.Contains(paletteText(m), "bubbles") {
		t.Fatal("Bubbles starts on")
	}
	m.perform(menubar.EditPacks)
	if m.flow.Modal() == nil {
		t.Fatal("the Packs dialog should open")
	}
	m.packs.SetEnabled(definition.FamilyBubbles, false)
	m.refreshPalette()
	if strings.Contains(paletteText(m), "bubbles") {
		t.Fatalf("a disabled pack must leave the palette:\n%s", paletteText(m))
	}
	if _, ok := m.cat.Get("bubbles.spinner"); !ok {
		t.Fatal("designs that use the pack must still resolve its components")
	}
}

func TestAboutNamesTheProvidersOfTheComponents(t *testing.T) {
	m := newShell(t)
	text := m.aboutText()
	for _, want := range []string{"Components from:", "Lip Gloss (", "Bubbles ("} {
		if !strings.Contains(text, want) {
			t.Errorf("about lacks %q:\n%s", want, text)
		}
	}
}
