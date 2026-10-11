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

// pagedApp has two pages: Home, with a button that goes to Settings, and
// Settings, with a button that goes back.
func pagedApp() design.Document {
	doc := design.NewDocument("My App", 60, 12)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Open settings", Rect: design.Rect{X: 2, Y: 2, W: 20, H: 3},
		Props: map[string]string{"title": "HOME PAGE"}, Goto: "Settings"})
	doc.Pages = []design.Page{
		{Name: "Home"},
		{Name: "Settings", Nodes: []design.Node{
			{ID: "n50", Component: "lipgloss.box", Name: "Back", Rect: design.Rect{X: 2, Y: 2, W: 20, H: 3},
				Props: map[string]string{"title": "SETTINGS PAGE"}, Goto: "Home"},
		}},
	}
	return doc
}

func TestAPagedDesignIsAnApplicationWithARouter(t *testing.T) {
	p := Generate(pagedApp(), standard.Default())
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	for _, name := range []string{"main.go", "go.mod", "my_app_app.go", "home_screen_contract.go", "settings_screen_view.go", "cuppa_screens.go"} {
		if _, ok := p.Files[name]; !ok {
			t.Errorf("no %s in %v", name, keys(p.Files))
		}
	}
	router := p.Files["my_app_app.go"]
	for _, want := range []string{"type MyAppPage string", `MyAppPage = "Home"`, "Page MyAppPage", "func MyApp(p MyAppProps", "case HomeGoToSettings:", "case SettingsGoToHome:", "Went: true"} {
		if !strings.Contains(router, want) {
			t.Errorf("router lacks %q:\n%s", want, router)
		}
	}
	if strings.Contains(router, DesignDirective) {
		t.Error("no source was given, so none is embedded")
	}
	withSource := GenerateWith(pagedApp(), standard.Default(), Options{Source: []byte("source bytes")})
	if !strings.Contains(withSource.Files["my_app_app.go"], DesignDirective) {
		t.Error("the source is embedded next to the router")
	}
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

const appCheck = `package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestApp(t *testing.T) {
	m := newModel()
	show := func() string { return ansi.Strip(m.View().Content) }
	if v := show(); !strings.Contains(v, "HOME PAGE") || strings.Contains(v, "SETTINGS PAGE") {
		t.Fatalf("at the start:\n%s", v)
	}
	next, _ := m.Update(tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft})
	m = next.(model)
	if m.props.Page != MyAppPageSettings {
		t.Fatalf("page = %q", m.props.Page)
	}
	if v := show(); !strings.Contains(v, "SETTINGS PAGE") || strings.Contains(v, "HOME PAGE") {
		t.Fatalf("after the click:\n%s", v)
	}
	next, _ = m.Update(tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft})
	m = next.(model)
	if m.props.Page != MyAppPageHome {
		t.Fatalf("page = %q", m.props.Page)
	}
}
`

func TestTheApplicationChangesPageWhenAComponentThatGoesThereIsClicked(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(pagedApp(), standard.Default())); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app_check_test.go"), []byte(appCheck), 0o644); err != nil {
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
