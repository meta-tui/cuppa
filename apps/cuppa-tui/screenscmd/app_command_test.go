package screenscmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// An application written by --app opens again from its main.go.
func TestAppWritesARunnableApplicationThatOpensBackFromItsMainGo(t *testing.T) {
	doc := design.NewDocument("Shop", 60, 12)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{X: 1, Y: 1, W: 20, H: 4}, Goto: "Cart"})
	doc.Pages = []design.Page{{Name: "Home"}, {Name: "Cart", Nodes: []design.Node{
		{ID: "n7", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 1, Y: 1, W: 10, H: 1}},
	}}}
	dir := t.TempDir()
	src, err := disk.Save(filepath.Join(dir, "shop"), doc)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "app")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{src, "-o", out, "--app"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	for _, name := range []string{"main.go", "go.mod", "shop_app.go", "home_screen_view.go", "cart_screen_view.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("no %s: %v", name, err)
		}
	}
	got, err := disk.Load(filepath.Join(out, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Shop" || len(got.Pages) != 2 || got.Pages[1].Name != "Cart" {
		t.Fatalf("opened %+v", got.PageNames())
	}
}
