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

// pressTest runs inside the generated program: it clicks the second tab and
// then the third section header, printing what the window shows each time.
const pressTest = `package main

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func show(m model) string {
	return strings.ReplaceAll(ansi.Strip(m.View().Content), "\n", "|")
}

func TestPress(t *testing.T) {
	m := newModel()
	fmt.Println("BEFORE=" + show(m))
	next, _ := m.Update(tea.MouseClickMsg{X: 14, Y: 1, Button: tea.MouseLeft})
	m = next.(model)
	fmt.Println("AFTER=" + show(m))
}
`

// A program exported from a design with a tab container needs no code to
// change page: pressing a tab shows the components of its page and hides the
// others.
func TestAProgramChangesPageWhenATabIsPressed(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	cat := standard.Default()
	doc := design.NewDocument("pages", 60, 12)
	tabs := doc.Add(design.Node{Component: "cuppa.tabs", Name: "Tabs", Rect: design.Rect{X: 2, Y: 1, W: 40, H: 8},
		Props: map[string]string{"pages": "Alpha,Beta", "active": "0"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "OnAlpha", In: tabs.ID, Page: 0,
		Rect: design.Rect{X: 4, Y: 4, W: 12, H: 1}, Props: map[string]string{"text": "ALPHA PAGE"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "OnBeta", In: tabs.ID, Page: 1,
		Rect: design.Rect{X: 4, Y: 4, W: 12, H: 1}, Props: map[string]string{"text": "BETA PAGE"}})
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(doc, cat)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "press_test.go"), []byte(pressTest), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	tidy := exec.Command(goBin, "mod", "tidy")
	tidy.Dir, tidy.Env = dir, env
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	run := exec.Command(goBin, "test", "-run", "TestPress", "-v", ".")
	run.Dir, run.Env = dir, env
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("go test: %v\n%s", err, out)
	}
	var before, after string
	for _, line := range strings.Split(string(out), "\n") {
		if rest, ok := strings.CutPrefix(line, "BEFORE="); ok {
			before = rest
		}
		if rest, ok := strings.CutPrefix(line, "AFTER="); ok {
			after = rest
		}
	}
	if !strings.Contains(before, "ALPHA PAGE") || strings.Contains(before, "BETA PAGE") {
		t.Errorf("at the start only page Alpha shows:\n%s", before)
	}
	if strings.Contains(after, "ALPHA PAGE") || !strings.Contains(after, "BETA PAGE") {
		t.Errorf("after pressing Beta only its page shows:\n%s", after)
	}
}
