package gosource

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	for _, name := range []string{"go.mod", "main.go", "app/app.go", "pages/home/home.cuppa.go", "pages/settings/settings.cuppa.go", "internal/cuppa/runtime.go", "internal/cuppa/screens.go"} {
		if _, ok := p.Files[name]; !ok {
			t.Errorf("no %s", name)
		}
	}
	page := p.Files["pages/home/home.cuppa.go"]
	for _, want := range []string{"package home", "type Props struct", "func Default() Props", "func View(p Props, w, h int) cuppa.Frame", "func Handle(f cuppa.Frame, msg tea.Msg) (Event, bool)", "type GoToSettingsEvent struct", `cuppa "example.com/shop/internal/cuppa"`} {
		if !strings.Contains(page, want) {
			t.Errorf("home.cuppa.go lacks %q:\n%s", want, page)
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
	if _, err := os.Stat(filepath.Join(dir, "pages", "settings", "settings.cuppa.go")); err == nil {
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

// The defaults of a page keep the names of the inputs: a layout input called
// like the page's own prefix (FolderRows on the page Folder) is a field, not a
// name of the page, so it is not shortened.
func TestADefaultKeepsTheNamesOfTheInputsInASlicePage(t *testing.T) {
	doc := design.NewDocument("Folder", 100, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Left", Rect: design.Rect{W: 10, H: 5}, Layout: design.Layout{W: "$FolderRows"}})
	doc.Samples = map[string]string{"FolderRows": "3"}
	p, err := GenerateSlices(doc, standard.Default(), Options{Module: "x"})
	if err != nil {
		t.Fatal(err)
	}
	src := strings.Join(strings.Fields(p.Files["pages/folder/folder.cuppa.go"]), " ")
	if !strings.Contains(src, "FolderRows: 3,") || strings.Contains(src, " Rows: 3,") {
		t.Errorf("Default does not keep the input name:\n%s", src)
	}
}

// A library export is packages for a module that exists: no go.mod, main.go or
// router, the runtime under internal/, a page in its own folder, and the
// runtime types each page names declared as aliases.
func TestALibraryExportIsPackagesWithoutAnApplication(t *testing.T) {
	doc := design.NewDocument("Setup", 100, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 5}})
	doc.Pages = []design.Page{{Name: "Setup", Folder: "setup/preferences"}, {Name: "Run"}}
	doc.Page = 0
	p, err := GenerateSlices(doc, standard.Default(), Options{Module: "app/libs/core", Library: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, gone := range []string{"go.mod", "main.go", "README.md", "app/app.go"} {
		if _, ok := p.Files[gone]; ok {
			t.Errorf("a library export wrote %s", gone)
		}
	}
	page, ok := p.Files["setup/preferences/preferences.cuppa.go"]
	if !ok {
		t.Fatalf("no page in its folder: %v", keysOf(p.Files))
	}
	for _, want := range []string{"package preferences", `"app/libs/core/internal/cuppa"`, "CellStyle = cuppa.CellStyle", "RowSelected = cuppa.RowSelected"} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	if _, ok := p.Files["pages/run/run.cuppa.go"]; !ok {
		t.Errorf("the other page is not in pages/run: %v", keysOf(p.Files))
	}
	if _, ok := p.Files["internal/cuppa/runtime.go"]; !ok {
		t.Errorf("no runtime under internal/cuppa: %v", keysOf(p.Files))
	}
	withRouter, err := GenerateSlices(doc, standard.Default(), Options{Module: "app/libs/core", Library: true, Router: true})
	if err != nil {
		t.Fatal(err)
	}
	router := withRouter.Files["app/app.go"]
	if !strings.Contains(router, `setup "app/libs/core/setup/preferences"`) {
		t.Errorf("the router does not import the page by its folder:\n%s", router)
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// --check of a library export fails on a hand-edited generated file and on a
// page that moved to another folder, and never looks at files that are not
// Cuppa's.
func TestCheckingALibraryExportFindsEditsAndMovedPages(t *testing.T) {
	doc := design.NewDocument("Setup", 100, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 5}})
	doc.Pages = []design.Page{{Name: "Setup", Folder: "setup/prefs"}, {Name: "Run"}}
	opts := Options{Module: "app/core", Library: true}
	p, err := GenerateSlices(doc, standard.Default(), opts)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	own := filepath.Join(dir, "engine.go")
	if err := os.WriteFile(own, []byte("package core\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceProject(dir, p); err != nil {
		t.Fatal(err)
	}
	if diffs := CheckProject(dir, p); len(diffs) != 0 {
		t.Fatalf("fresh export is not current: %v", diffs)
	}
	page := filepath.Join(dir, "setup", "prefs", "prefs.cuppa.go")
	if err := os.WriteFile(page, []byte("// Code generated by Cuppa. DO NOT EDIT.\n\npackage prefs\n// edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if diffs := CheckProject(dir, p); len(diffs) != 1 || !strings.Contains(diffs[0], "prefs.cuppa.go") {
		t.Fatalf("an edited file: %v", diffs)
	}
	doc.Pages[0].Folder = "setup/preferences"
	moved, err := GenerateSlices(doc, standard.Default(), opts)
	if err != nil {
		t.Fatal(err)
	}
	diffs := strings.Join(CheckProject(dir, moved), "\n")
	if !strings.Contains(diffs, "preferences.cuppa.go") || !strings.Contains(diffs, "prefs.cuppa.go") || strings.Contains(diffs, "engine.go") {
		t.Fatalf("a moved page: %v", diffs)
	}
	if err := ReplaceProject(dir, moved); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(page); err == nil {
		t.Error("the page's old file is still there")
	}
	if _, err := os.Stat(own); err != nil {
		t.Error("the person's own file was removed")
	}
}

// The master is not a page of its own: each page's package draws it, and its
// inputs are the page's inputs under the same names.
func TestTheMasterIsOnEveryPageAndIsNotAPage(t *testing.T) {
	doc := design.NewDocument("Shop", 80, 20)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Home box", Rect: design.Rect{X: 2, Y: 3, W: 10, H: 4}})
	title := design.Node{ID: "n9", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{W: 80, H: 1},
		Props: map[string]string{"text": "Shop"}, Bind: map[string]string{"text": "Title"}}
	doc.Pages = []design.Page{
		{Name: "Home"},
		{Name: "Cart", Nodes: []design.Node{{ID: "n2", Component: "lipgloss.box", Name: "Cart box", Rect: design.Rect{X: 2, Y: 3, W: 10, H: 4}}}},
		{Name: "Master", Master: true, Nodes: []design.Node{title}},
	}
	doc.PerPage = true
	p, err := GenerateSlices(doc, standard.Default(), Options{Module: "shop", Library: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.Files["pages/master/master.cuppa.go"]; ok {
		t.Error("the master was exported as a page")
	}
	for _, page := range []string{"pages/home/home.cuppa.go", "pages/cart/cart.cuppa.go"} {
		src := p.Files[page]
		if !strings.Contains(src, "Title string") || !strings.Contains(src, `"Title"`) {
			t.Errorf("%s does not have the master's Title input", page)
		}
	}
}
