package gosource

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// pagesScreen has a tab container with a label on each page and a panel stack
// with a label in its second section.
func pagesScreen() design.Document {
	doc := design.NewDocument("Pages", 60, 24)
	tabs := doc.Add(design.Node{Component: "cuppa.tabs", Name: "Tabs", Rect: design.Rect{X: 1, Y: 0, W: 40, H: 6},
		Props: map[string]string{"pages": "Alpha,Beta", "active": "0"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "OnAlpha", In: tabs.ID, Page: 0,
		Rect: design.Rect{X: 3, Y: 3, W: 12, H: 1}, Props: map[string]string{"text": "ALPHA PAGE"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "OnBeta", In: tabs.ID, Page: 1,
		Rect: design.Rect{X: 3, Y: 3, W: 12, H: 1}, Props: map[string]string{"text": "BETA PAGE"}})
	panels := doc.Add(design.Node{Component: "cuppa.panels", Name: "Stack", Rect: design.Rect{X: 1, Y: 8, W: 40, H: 12},
		Props: map[string]string{"pages": "One,Two,Three", "open": "0,1"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "InTwo", In: panels.ID, Page: 1,
		Rect: design.Rect{X: 3, Y: 14, W: 12, H: 1}, Props: map[string]string{"text": "SECTION TWO"}})
	return doc
}

func TestContainersBecomeInputsInTheContract(t *testing.T) {
	p := GenerateScreens([]design.Document{pagesScreen()}, standard.Default(), "screens")
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	c := p.Files["pages_screen_contract.go"]
	for _, want := range []string{"TabsActive int", "StackOpen []int", "TabsActive: 0", "[]int{0, 1}"} {
		if !strings.Contains(c, want) {
			t.Errorf("contract lacks %q:\n%s", want, c)
		}
	}
	if v := p.Files["pages_screen_view.go"]; !strings.Contains(v, "func PagesPress(p *PagesProps, f Frame, msg tea.Msg) bool") {
		t.Errorf("no press handler:\n%s", v)
	}
}

const pagesCheck = `package screens

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func plain(p PagesProps) string { return ansi.Strip(Pages(p, 60, 24).View) }

func TestPages(t *testing.T) {
	p := DefaultPagesProps()
	v := plain(p)
	if !strings.Contains(v, "ALPHA PAGE") || strings.Contains(v, "BETA PAGE") || !strings.Contains(v, "SECTION TWO") {
		t.Fatalf("at the start:\n%s", v)
	}
	// Pressing the Beta tab (the bar is the first row of the container).
	f := Pages(p, 60, 24)
	if !PagesPress(&p, f, tea.MouseClickMsg{X: 11, Y: 0, Button: tea.MouseLeft}) {
		t.Fatal("the press on the tab was not handled")
	}
	if p.TabsActive != 1 {
		t.Fatalf("active = %d", p.TabsActive)
	}
	v = plain(p)
	if strings.Contains(v, "ALPHA PAGE") || !strings.Contains(v, "BETA PAGE") {
		t.Fatalf("after the press:\n%s", v)
	}
	// Closing section Two hides its label.
	f = Pages(p, 60, 24)
	var hdr int
	for y, line := range strings.Split(ansi.Strip(f.View), "\n") {
		if strings.Contains(line, "▾ Two") {
			hdr = y
		}
	}
	if !PagesPress(&p, f, tea.MouseClickMsg{X: 3, Y: hdr, Button: tea.MouseLeft}) {
		t.Fatalf("the press on the header at row %d was not handled", hdr)
	}
	if strings.Contains(plain(p), "SECTION TWO") {
		t.Errorf("section Two is closed but its label shows:\n%s", plain(p))
	}
	if len(p.StackOpen) != 1 || p.StackOpen[0] != 0 {
		t.Errorf("open = %v", p.StackOpen)
	}
}
`

func TestContainersScreenBuildsAndChangesPage(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	p := GenerateScreens([]design.Document{pagesScreen()}, standard.Default(), "screens")
	if err := WriteScreens(dir, p); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages_check_test.go"), []byte(pagesCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"mod", "tidy"}, {"vet", "./..."}, {"test", "./..."}} {
		cmd := exec.Command(goBin, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}
