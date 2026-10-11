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

func TestTheSliceLayoutHasAFolderAndAPackagePerPage(t *testing.T) {
	p, err := GenerateSlices(pagedApp(), standard.Default(), Options{Module: "example.com/shop"})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "main.go", "app/app.go", "pages/home/home.go", "pages/settings/settings.go", "internal/cuppa/runtime.go", "internal/cuppa/screens.go"} {
		if _, ok := p.Files[name]; !ok {
			t.Errorf("no %s", name)
		}
	}
	page := p.Files["pages/home/home.go"]
	for _, want := range []string{"package home", "type Props struct", "func Default() Props", "func View(p Props, w, h int) cuppa.Frame", "func Handle(f cuppa.Frame, msg tea.Msg) (Event, bool)", "type GoToSettingsEvent struct", `cuppa "example.com/shop/internal/cuppa"`} {
		if !strings.Contains(page, want) {
			t.Errorf("home.go lacks %q:\n%s", want, page)
		}
	}
	if strings.Contains(page, "type HomeProps") || strings.Contains(page, "var homeParts") {
		t.Errorf("the page keeps its prefix:\n%s", page)
	}
	router := p.Files["app/app.go"]
	for _, want := range []string{"package app", `"example.com/shop/pages/home"`, "Home home.Props", "case home.GoToSettingsEvent:", "func Handle(p *Props"} {
		if !strings.Contains(router, want) {
			t.Errorf("app.go lacks %q:\n%s", want, router)
		}
	}
	if !strings.Contains(p.Files["internal/cuppa/runtime.go"], "package cuppa") {
		t.Error("the runtime is package cuppa")
	}
}

const sliceCheck = `package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestShop(t *testing.T) {
	p := Default()
	v := ansi.Strip(View(p, 60, 12).View)
	if !strings.Contains(v, "HOME PAGE") || strings.Contains(v, "SETTINGS PAGE") {
		t.Fatalf("at the start:\n%s", v)
	}
	f := View(p, 60, 12)
	ev, ok := Handle(&p, f, tea.MouseClickMsg{X: 5, Y: 3, Button: tea.MouseLeft})
	if !ok || !ev.Went || p.Page != PageSettings {
		t.Fatalf("event %+v page %q", ev, p.Page)
	}
	v = ansi.Strip(View(p, 60, 12).View)
	if !strings.Contains(v, "SETTINGS PAGE") {
		t.Fatalf("after the click:\n%s", v)
	}
}
`

// The slice layout builds, and a click that goes to a page changes the page.
func TestTheSliceLayoutBuildsAndRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	p, err := GenerateSlices(pagedApp(), standard.Default(), Options{Module: "example.com/shop"})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	dir := filepath.Join(t.TempDir(), "shop")
	if err := Write(dir, p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app", "app_check_test.go"), []byte(sliceCheck), 0o644); err != nil {
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

var _ = design.Document{}

// Exporting again replaces what Cuppa generated, removes the page that is
// gone and leaves the person's own files and go.mod alone.
func TestReplacingAProjectKeepsWhatIsTheOwnersAndRemovesGonePages(t *testing.T) {
	cat := standard.Default()
	dir := t.TempDir()
	first, err := GenerateSlices(pagedApp(), cat, Options{Module: "example.com/shop"})
	if err != nil {
		t.Fatal(err)
	}
	if err := ReplaceProject(dir, first); err != nil {
		t.Fatal(err)
	}
	own := filepath.Join(dir, "pages", "home", "update.go")
	if err := os.WriteFile(own, []byte("package home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module example.com/shop\n\ngo 1.24\n// mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if diffs := CheckProject(dir, first); len(diffs) != 0 {
		t.Fatalf("a fresh export is current: %v", diffs)
	}
	// The page Settings goes away.
	doc := pagedApp()
	doc.Pages = doc.Pages[:1]
	doc.Page = 0
	second, err := GenerateSlices(doc, cat, Options{Module: "example.com/shop"})
	if err != nil {
		t.Fatal(err)
	}
	if diffs := CheckProject(dir, second); len(diffs) == 0 {
		t.Fatal("a removed page is a difference")
	}
	if err := ReplaceProject(dir, second); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "settings", "settings.go")); err == nil {
		t.Error("the page that is gone is still there")
	}
	for _, kept := range []string{own, gomod} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("%s was removed", kept)
		}
	}
	if data, _ := os.ReadFile(gomod); !strings.Contains(string(data), "// mine") {
		t.Error("go.mod was overwritten")
	}
	if diffs := CheckProject(dir, second); len(diffs) != 0 {
		t.Errorf("after replacing: %v", diffs)
	}
}

const sliceContainersCheck = `package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPagesInSlices(t *testing.T) {
	p := Default()
	show := func() string { return ansi.Strip(View(p, 60, 24).View) }
	if v := show(); !strings.Contains(v, "ALPHA PAGE") || strings.Contains(v, "BETA PAGE") {
		t.Fatalf("at the start:\n%s", v)
	}
	if !Press(&p, View(p, 60, 24), tea.MouseClickMsg{X: 11, Y: 0, Button: tea.MouseLeft}) {
		t.Fatal("the press on the tab was not handled")
	}
	if v := show(); strings.Contains(v, "ALPHA PAGE") || !strings.Contains(v, "BETA PAGE") {
		t.Fatalf("after the press:\n%s", v)
	}
}
`

// Tabs and panels work in a page package: the input is in the page's props and
// the router's Press reaches it.
func TestContainersWorkInASlicePage(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	p, err := GenerateSlices(pagesScreen(), standard.Default(), Options{Module: "example.com/pages"})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "pages")
	if err := Write(dir, p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app", "containers_check_test.go"), []byte(sliceContainersCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	goRun(t, dir, "mod", "tidy")
	goRun(t, dir, "vet", "./...")
	goRun(t, dir, "test", "./...")
}
