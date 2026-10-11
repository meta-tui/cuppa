package gosource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// masterDesign is a shop with a title and a key bar on the master; Cart has
// no key bar and its own text for the title.
func masterDesign() design.Document {
	doc := design.NewDocument("Shop", 60, 12)
	doc.Add(design.Node{ID: "h1", Component: "lipgloss.box", Name: "Home box", Rect: design.Rect{X: 2, Y: 3, W: 10, H: 4}})
	title := design.Node{ID: "t", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{W: 60, H: 1},
		Props: map[string]string{"text": "Shop"}, Bind: map[string]string{"text": "Title"}}
	keys := design.Node{ID: "k", Component: "lipgloss.keybar", Name: "Key bar", Rect: design.Rect{Y: 11, W: 60, H: 1},
		Props: map[string]string{"hints": "q:quit,enter:buy", "gap": "2"}, Bind: map[string]string{"hints": "Hints"},
		ShowIf: "Show keys", Event: "Hint", Layout: design.Layout{Y: "100% - 1", W: "100%"}}
	doc.Pages = []design.Page{
		{Name: "Home"},
		{Name: "Cart", Hide: []design.NodeID{"k"}, Values: map[string]string{"Title": "Cart"},
			Nodes: []design.Node{{ID: "c1", Component: "lipgloss.box", Name: "Cart box", Rect: design.Rect{X: 2, Y: 3, W: 10, H: 4}}}},
		{Name: "Master", Master: true, Nodes: []design.Node{title, keys}},
	}
	doc.PerPage = true
	return doc
}

// The master's parts are written once, in a package of their own, and a page
// takes them from it with its own props (ADR 0015, step 2).
func TestTheMasterIsWrittenOnceInTheSliceLayout(t *testing.T) {
	p, err := GenerateSlices(masterDesign(), standard.Default(), Options{Module: "shop", Library: true})
	if err != nil {
		t.Fatal(err)
	}
	src, ok := p.Files["internal/components/master/master.go"]
	if !ok {
		t.Fatalf("no master package; files: %v", fileNames(p))
	}
	if strings.Count(src, `"lipgloss.keybar"`) != 1 {
		t.Error("the master's package does not hold the key bar once")
	}
	for _, page := range []string{"pages/home/home.go", "pages/cart/cart.go"} {
		page := p.Files[page]
		if strings.Contains(page, "lipgloss.keybar") || strings.Contains(page, `Name: "Title"`) {
			t.Errorf("a page still draws the master's parts itself:\n%s", page)
		}
		if !strings.Contains(page, "master.PartsFor(") {
			t.Errorf("a page does not take the master's parts from the package")
		}
	}
	cart := p.Files["pages/cart/cart.go"]
	if !strings.Contains(cart, `master.PartsFor(masterProps, "Key bar")`) {
		t.Errorf("Cart does not hide the key bar:\n%s", cart)
	}
	if !strings.Contains(cart, `Title: "Cart"`) {
		t.Errorf("Cart does not start from its own text for the title")
	}
	// The page's API is unchanged.
	home := p.Files["pages/home/home.go"]
	for _, want := range []string{"func View(", "func Default()", "func Layout(", "func Handle(", "func HintsHintAt(p Props, x int)", "type HintsHint = master.HintsHint"} {
		if !strings.Contains(home, want) {
			t.Errorf("Home has no %q", want)
		}
	}
}

// The pages draw what they drew when each carried the master's parts.
func TestTheSharedMasterDrawsWhatTheCopiesDrew(t *testing.T) {
	if testing.Short() {
		t.Skip("builds projects with the go tool")
	}
	doc := masterDesign()
	cat := standard.Default()
	dump := func(pkg string, calls ...string) string {
		var b strings.Builder
		b.WriteString("package " + pkg + "\n\nimport (\n\t\"fmt\"\n\t\"testing\"\n\n\t\"github.com/charmbracelet/x/ansi\"\n)\n\nfunc TestDump(t *testing.T) {\n")
		for _, c := range calls {
			b.WriteString(c)
		}
		b.WriteString("}\n")
		return b.String()
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	// What the copies drew: the flat layout writes the master into each page.
	flat := GenerateScreens(doc.Screens(), cat, "screens")
	flatDir := filepath.Join(t.TempDir(), "flat")
	must(WriteScreens(flatDir, flat))
	gomod := func(name string) string {
		return "module " + name + "\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	}
	var calls []string
	for _, name := range []string{"Home", "Cart"} {
		for _, w := range []int{60, 40} {
			calls = append(calls, fmt.Sprintf("\tfmt.Println(\"DRAW %s %d<<\" + ansi.Strip(%s(Default%sProps(), %d, 12).View) + \">>DRAW\")\n", name, w, name, name, w))
		}
	}
	must(os.WriteFile(filepath.Join(flatDir, "go.mod"), []byte(gomod("example.com/flat")), 0o644))
	must(os.WriteFile(filepath.Join(flatDir, "dump_test.go"), []byte(dump("screens", calls...)), 0o644))
	goRun(t, flatDir, "mod", "tidy")
	want := goRun(t, flatDir, "test", "-run", "TestDump", "-v", ".")

	// What the shared master draws: each page package on its own.
	p, err := GenerateSlices(doc, cat, Options{Module: "example.com/sliced", Library: true})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "sliced")
	must(Write(dir, p))
	must(os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod("example.com/sliced")), 0o644))
	var got strings.Builder
	for _, name := range []string{"Home", "Cart"} {
		pkg := strings.ToLower(name)
		var c []string
		for _, w := range []int{60, 40} {
			c = append(c, fmt.Sprintf("\tfmt.Println(\"DRAW %s %d<<\" + ansi.Strip(View(Default(), %d, 12).View) + \">>DRAW\")\n", name, w, w))
		}
		must(os.WriteFile(filepath.Join(dir, "pages", pkg, "dump_test.go"), []byte(dump(pkg, c...)), 0o644))
	}
	goRun(t, dir, "mod", "tidy")
	goRun(t, dir, "vet", "./...")
	for _, name := range []string{"home", "cart"} {
		got.WriteString(goRun(t, dir, "test", "-run", "TestDump", "-v", "./pages/"+name))
	}
	for _, name := range []string{"Home", "Cart"} {
		for _, w := range []int{60, 40} {
			tag := fmt.Sprintf("DRAW %s %d<<", name, w)
			a, b := between2(want, tag), between2(got.String(), tag)
			if a == "" || a != b {
				t.Errorf("%s at %d columns draws differently:\n-- copies\n%s\n-- shared\n%s", name, w, a, b)
			}
		}
	}
}

func between2(out, tag string) string {
	_, rest, ok := strings.Cut(out, tag)
	if !ok {
		return ""
	}
	drawn, _, _ := strings.Cut(rest, ">>DRAW")
	return drawn
}

func fileNames(p Project) []string {
	var out []string
	for n := range p.Files {
		out = append(out, n)
	}
	return out
}
