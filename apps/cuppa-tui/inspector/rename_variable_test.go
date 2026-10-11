package inspector

import (
	"strings"
	"testing"

	"github.com/meta-tui/cuppa/libs/document/design"
)

func designID(s string) design.NodeID { return design.NodeID(s) }

// answerWith binds an asker that records the question and presses a button.
func answerWith(m *Model, button string, asked *[]string) {
	m.BindAsker(func(title, body string, buttons []string, then func(string)) {
		*asked = append(*asked, title+": "+body+" ["+strings.Join(buttons, "|")+"]")
		then(button)
	})
}

func twoLabelsSharingAColour(t *testing.T) (*Model, func(), string, string) {
	t.Helper()
	m, ed, a := setup(t, "lipgloss.label")
	b, err := ed.Add("lipgloss.label", 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := ed.SetColour(a, "color", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	if err := ed.UseSwatch(b, "color", "Foreground"); err != nil {
		t.Fatal(err)
	}
	ed.Select(a)
	m.Lines()
	rename := func() {
		clickText(t, m, "name", "Foreground")
		clickText(t, m, "Rename", "Rename")
		for range "Foreground" {
			m.Key("", true, false, false)
		}
		for _, r := range "Accent" {
			m.Key(string(r), false, false, false)
		}
		m.Key("", false, true, false)
	}
	return m, rename, string(a), string(b)
}

func TestRenamingAVariableUsedElsewhereAsksAndCanRenameEverywhere(t *testing.T) {
	m, rename, a, b := twoLabelsSharingAColour(t)
	var asked []string
	answerWith(m, "Rename everywhere", &asked)
	rename()
	if len(asked) != 1 || !strings.Contains(asked[0], `"Foreground" is used in 2 places`) || !strings.Contains(asked[0], "Rename everywhere|Only this one|Cancel") {
		t.Fatalf("asked %q", asked)
	}
	doc := m.ed.Document()
	na, _ := doc.Get(designID(a))
	nb, _ := doc.Get(designID(b))
	if na.Props["color"] != "@Accent" || nb.Props["color"] != "@Accent" {
		t.Fatalf("a %q b %q", na.Props["color"], nb.Props["color"])
	}
	if _, ok := doc.Theme.Colour("Foreground"); ok {
		t.Error("old name kept")
	}
}

func TestRenamingAVariableCanChangeOnlyThisUse(t *testing.T) {
	m, rename, a, b := twoLabelsSharingAColour(t)
	var asked []string
	answerWith(m, "Only this one", &asked)
	rename()
	doc := m.ed.Document()
	na, _ := doc.Get(designID(a))
	nb, _ := doc.Get(designID(b))
	if na.Props["color"] != "@Accent" || nb.Props["color"] != "@Foreground" {
		t.Fatalf("a %q b %q", na.Props["color"], nb.Props["color"])
	}
	if c, _ := doc.Theme.Colour("Accent"); c != "#ff0000" {
		t.Errorf("the new variable starts with the colour of the old one, got %q", c)
	}
}

func TestCancellingTheQuestionChangesNothing(t *testing.T) {
	m, rename, a, _ := twoLabelsSharingAColour(t)
	var asked []string
	answerWith(m, "Cancel", &asked)
	rename()
	na, _ := m.ed.Document().Get(designID(a))
	if na.Props["color"] != "@Foreground" {
		t.Fatalf("a %q", na.Props["color"])
	}
}

func TestRenamingAVariableUsedOnlyHereJustRenamesIt(t *testing.T) {
	m, ed, a := setup(t, "lipgloss.label")
	if err := ed.SetColour(a, "color", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	m.Lines()
	var asked []string
	answerWith(m, "Cancel", &asked)
	clickText(t, m, "name", "Foreground")
	clickText(t, m, "Rename", "Rename")
	for range "Foreground" {
		m.Key("", true, false, false)
	}
	for _, r := range "Accent" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	if len(asked) != 0 {
		t.Fatalf("asked %q for a variable used once", asked)
	}
	if _, ok := ed.Document().Theme.Colour("Accent"); !ok {
		t.Fatal("not renamed")
	}
}
