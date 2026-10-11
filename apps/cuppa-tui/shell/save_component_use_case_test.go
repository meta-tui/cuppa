package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

func TestSavingTheSelectionAsAComponentMakesItAvailableEverywhere(t *testing.T) {
	m, dir := packsShell(t)
	frame, _ := m.ed.Add("lipgloss.box", 4, 3)
	title, _ := m.ed.Add("lipgloss.label", 6, 4)
	if err := m.ed.SetProp(title, "text", "Matcha"); err != nil {
		t.Fatal(err)
	}
	m.ed.Select(frame, title)
	if !m.canSaveComponent() {
		t.Fatal("two unlocked components can be saved")
	}

	m.perform(menubar.EditComponent)
	dlg := m.flow.Modal()
	if dlg == nil {
		t.Fatal("the name prompt should open")
	}
	for range "Group 1" {
		dlg.Key("", true, false, false)
	}
	dlg.Key("Tea Card", false, false, false)
	dlg.Key("", false, true, false)
	m.flow.Resolve()

	if _, err := os.Stat(filepath.Join(dir, "my-components.cupp")); err != nil {
		t.Fatalf("the personal pack file is written: %v", err)
	}
	def, ok := m.cat.Get("my-components.tea-card")
	if !ok || len(def.Props) == 0 {
		t.Fatalf("the component is in the catalog with properties: %+v", def)
	}
	if n := m.flow.Modal(); n == nil {
		t.Fatal("a notice says it was saved")
	}

	// A second one in the same pack gets its own id.
	m.flow.Modal().Key("", false, true, false)
	m.flow.Resolve()
	g, _ := m.ed.Primary()
	if _, err := m.storeComponent(g, "Tea Card"); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.cat.Get("my-components.tea-card-2"); !ok {
		t.Fatal("same name, next id")
	}

	// Place it in a fresh design and draw it.
	m.ed.Clear()
	id, err := m.ed.Add("my-components.tea-card", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	g2 := scene.Render(m.ed.Document(), m.cat)
	var out strings.Builder
	for y := 0; y < g2.H; y++ {
		for x := 0; x < g2.W; x++ {
			if ch := g2.At(x, y).Ch; ch != 0 {
				out.WriteRune(ch)
			}
		}
	}
	if !strings.Contains(out.String(), "Matcha") {
		t.Fatalf("the placed component draws its parts (%s):\n%s", id, out.String())
	}
}

func TestSaveAsComponentNeedsASelection(t *testing.T) {
	m, _ := packsShell(t)
	m.perform(menubar.EditComponent)
	if m.flow.Modal() == nil {
		t.Fatal("a notice explains what to select")
	}
}
