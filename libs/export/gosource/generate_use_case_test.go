package gosource

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// richDesign uses every component that has real code, one that does not, a
// group, and a component from a pack.
func richDesign() (design.Document, Catalog) {
	pack := cupp.Pack{ID: "tea-shop", Name: "Tea shop", Components: []design.Composite{{
		ID: "card", Name: "Card", W: 20, H: 5,
		Nodes: []design.Node{
			{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}},
			{ID: "b", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 2, Y: 1, W: 10, H: 1}},
		},
		Props: []design.Exposed{{Key: "title", Label: "Title", Kind: "text", Default: "Tea", Target: "b", TargetProp: "text"}},
	}}}
	cat, _ := cupp.Extend(standard.Default(), []cupp.Pack{pack})

	doc := design.NewDocument("My Tea \"Shop\" #1", 120, 40)
	doc.Background = "#102030"
	at := 0
	for _, id := range []string{
		"lipgloss.box", "lipgloss.label", "bubbles.textinput", "bubbles.textarea", "bubbles.list", "bubbles.table",
		"bubbles.viewport", "bubbles.paginator", "bubbles.spinner", "bubbles.progress", "bubbles.stopwatch", "bubbles.timer",
		"lipgloss.list", "bubbles.tree", "huh.spinner",
		"community.frame", "community.dialog", "community.statusmessage", "community.toast",
		"community.bigtext", "community.qrcode", "community.image",
		"lipgloss.table", "lipgloss.tree", "lipgloss.tabs", "lipgloss.joinh", "lipgloss.joinv", "lipgloss.place",
		"bubbles.help", "bubbles.filepicker",
		"community.flexbox", "community.boxer", "community.datepicker", "community.overlay", "community.statusbar", "community.filetree",
		"huh.input", "huh.text", "huh.select", "huh.multiselect", "huh.confirm", "huh.note", "huh.filepicker", "huh.form",
		"glamour.markdown",
		"ntcharts.sparkline", "ntcharts.barchart", "ntcharts.linechart", "ntcharts.streamline", "ntcharts.timeseries", "ntcharts.heatmap", "ntcharts.canvas",
		"community.bubbletable",
	} {
		def, _ := cat.Get(id)
		doc.Add(design.Node{Component: id, Name: def.Name, Rect: design.Rect{X: (at % 4) * 28, Y: (at / 4) * 8, W: def.DefaultSize.W, H: def.DefaultSize.H}})
		at++
	}
	doc.Add(design.Node{Component: "tea-shop.card", Name: "Card", Rect: design.Rect{X: 80, Y: 30, W: 30, H: 6},
		Props: map[string]string{"title": "Matcha \"latte\""}})
	doc.Add(design.Node{Component: design.GroupComponent, Name: "Group", Rect: design.Rect{X: 2, Y: 30, W: 40, H: 8}, BaseW: 20, BaseH: 4,
		Children: []design.Node{
			{ID: "g1", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 4}},
			{ID: "g2", Component: "lipgloss.label", Name: "Text", Rect: design.Rect{X: 1, Y: 1, W: 10, H: 1}},
		}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Fade", Rect: design.Rect{X: 50, Y: 30, W: 24, H: 5},
		Props: map[string]string{"title": "Fade", "color": "#ff5fd7", "gradient": "#5f87ff"}})
	// The drawing: a few painted cells, generated as runs.
	drawn := drawlayer.New()
	drawn.Paint(drawlayer.Cell{X: 1, Y: 1, Ch: '╭', Fg: "212"}, drawlayer.Cell{X: 2, Y: 1, Ch: '─', Fg: "212"}, drawlayer.Cell{X: 3, Y: 1, Ch: '╮', Fg: "212"})
	doc.Add(design.Node{Component: drawlayer.Component, Name: "Drawing", Rect: design.Rect{W: 120, H: 40},
		Props: map[string]string{drawlayer.PropCells: drawn.Encode()}})
	// A component the generator does not know: drawn as an empty frame and said so.
	doc.Add(design.Node{Component: "mystery.widget", Name: "Mystery", Rect: design.Rect{X: 100, Y: 30, W: 10, H: 4}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Hidden", Rect: design.Rect{W: 5, H: 3}, Hidden: true})
	return doc, cat
}

func TestTheProjectHasTheFilesAndTheLayoutOfTheDesign(t *testing.T) {
	doc, cat := richDesign()
	p := Generate(doc, cat)
	for _, name := range []string{"go.mod", "main.go", "runtime.go", "layout.go", "README.md"} {
		if p.Files[name] == "" {
			t.Errorf("missing %s", name)
		}
	}
	if p.Module != "my-tea-shop-1" || !strings.Contains(p.Files["go.mod"], "module my-tea-shop-1") {
		t.Fatalf("module = %q", p.Module)
	}
	layout := p.Files["layout.go"]
	for _, want := range []string{
		"canvasW    = 120", `background = "#102030"`, `Kind: "bubbles.textinput"`,
		`"text": "Matcha \"latte\""`, // the pack component's exposed value reaches its part
	} {
		if !strings.Contains(layout, want) {
			t.Errorf("layout.go lacks %s", want)
		}
	}
	if strings.Contains(layout, "tea-shop.card") || strings.Contains(layout, "cuppa.group") {
		t.Error("groups and pack components are expanded into their parts")
	}
	if strings.Contains(layout, `Name: "Hidden"`) {
		t.Error("hidden components are not generated")
	}
	if len(p.Notes) != 1 || !strings.Contains(p.Notes[0], "mystery.widget") || !strings.Contains(p.Files["README.md"], "mystery.widget") {
		t.Fatalf("what is not generated is said: %v", p.Notes)
	}
}

func TestGroupPartsAreScaledAndPlacedAtTheGroup(t *testing.T) {
	doc, cat := richDesign()
	layout := Generate(doc, cat).Files["layout.go"]
	// The group is 40 wide for a 20-wide base: its frame spans 40 cells from x=2.
	if !strings.Contains(layout, `Name: "Frame", X: 2, Y: 30, W: 40, H: 8`) {
		t.Fatalf("the grouped frame is scaled to the group:\n%s", layout)
	}
}

func TestTheSameDesignGivesTheSameFiles(t *testing.T) {
	doc, cat := richDesign()
	a, b := Generate(doc, cat), Generate(doc, cat)
	for name := range a.Files {
		if a.Files[name] != b.Files[name] {
			t.Fatalf("%s differs between runs", name)
		}
	}
}

func TestComponentsThatNeedALibraryGetTheirOwnFileAndModule(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("media", 80, 24)
	plain := Generate(doc, cat)
	if plain.Files["bigtext.go"] != "" || strings.Contains(plain.Files["go.mod"], "go-figure") {
		t.Fatal("a design without big text must not ask for go-figure")
	}
	doc.Add(design.Node{Component: "community.bigtext", Name: "Title", Rect: design.Rect{W: 40, H: 6}})
	doc.Add(design.Node{Component: "community.qrcode", Name: "Code", Rect: design.Rect{W: 29, H: 15}})
	doc.Add(design.Node{Component: "community.image", Name: "Photo", Rect: design.Rect{W: 20, H: 8}})
	p := Generate(doc, cat)
	for _, name := range []string{"bigtext.go", "qrcode.go", "image.go"} {
		if p.Files[name] == "" {
			t.Errorf("missing %s", name)
		}
	}
	for _, want := range []string{"github.com/common-nighthawk/go-figure", "github.com/skip2/go-qrcode"} {
		if !strings.Contains(p.Files["go.mod"], want) {
			t.Errorf("go.mod does not require %s:\n%s", want, p.Files["go.mod"])
		}
	}
	if len(p.Notes) != 0 {
		t.Errorf("every placed component is generated, notes = %v", p.Notes)
	}
}

func TestWriteRefusesAFolderThatAlreadyHoldsAProject(t *testing.T) {
	doc, cat := richDesign()
	p := Generate(doc, cat)
	dir := filepath.Join(t.TempDir(), "out")
	if err := Write(dir, p); err != nil {
		t.Fatal(err)
	}
	if err := Write(dir, p); err == nil {
		t.Fatal("a second export into the same folder must not overwrite")
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{"Untitled": "untitled", "  Hello, World!": "hello-world", "": "cuppa-design", "???": "cuppa-design"} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestTheGeneratedProjectCompiles is the point of the exporter: what it writes
// builds and passes go vet. It needs the go tool and the module proxy (or a
// warm module cache), so it is skipped in -short runs.
func TestTheGeneratedProjectCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	doc, cat := richDesign()
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(doc, cat)); err != nil {
		t.Fatal(err)
	}
	// The generated program must also fit every component in its rectangle.
	fit := "package main\n\nimport (\n\t\"testing\"\n\n\t\"charm.land/lipgloss/v2\"\n)\n\n" +
		"func TestEveryPartFitsItsRectangle(t *testing.T) {\n\tfor _, c := range layout {\n" +
		"\t\tv := build(c).view()\n" +
		"\t\tif lipgloss.Width(v) > c.W || lipgloss.Height(v) > c.H {\n" +
		"\t\t\tt.Errorf(\"%s (%s) draws %dx%d in a %dx%d rectangle\", c.Name, c.Kind, lipgloss.Width(v), lipgloss.Height(v), c.W, c.H)\n" +
		"\t\t}\n\t}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "fit_test.go"), []byte(fit), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"mod", "tidy"}, {"vet", "./..."}, {"build", "./..."}, {"test", "./..."}} {
		cmd := exec.Command(goBin, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

func TestGeneratedProgramCarriesTheThemeColoursAndTheOverrides(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("themed", 60, 20)
	doc.Theme = design.Theme{Border: "#336699", Muted: "#445566"}
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Follows", Rect: design.Rect{W: 10, H: 3}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Own", Rect: design.Rect{X: 20, W: 10, H: 3},
		Props: map[string]string{"color": "212"}})
	layout := Generate(doc, cat).Files["layout.go"]
	follows := layout[strings.Index(layout, `Name: "Follows"`):strings.Index(layout, `Name: "Own"`)]
	if !strings.Contains(follows, `"color": "#336699"`) {
		t.Errorf("a box with no colour follows the theme:\n%s", follows)
	}
	own := layout[strings.Index(layout, `Name: "Own"`):]
	if !strings.Contains(own, `"color": "212"`) {
		t.Errorf("an override stays:\n%s", own)
	}
	if !strings.Contains(layout, `"theme.muted": "#445566"`) {
		t.Errorf("the muted colour reaches the widgets:\n%s", layout)
	}
}
