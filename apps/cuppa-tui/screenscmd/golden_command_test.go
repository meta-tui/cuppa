package screenscmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func goldenDesign(t *testing.T) string {
	t.Helper()
	doc := design.NewDocument("Shop", 40, 6)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{X: 1, Y: 1, W: 20, H: 3},
		Props: map[string]string{"title": "HOME"}})
	doc.Pages = []design.Page{{Name: "Home"}, {Name: "Cart", Nodes: []design.Node{
		{ID: "n50", Component: "lipgloss.box", Name: "Cart box", Rect: design.Rect{X: 2, Y: 1, W: 18, H: 3}, Props: map[string]string{"title": "CART"}},
	}}}
	file, err := disk.Save(filepath.Join(t.TempDir(), "shop"), doc)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestGoldenDrawingsAreWrittenAndChecked(t *testing.T) {
	file := goldenDesign(t)
	dir := filepath.Join(t.TempDir(), "golden")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{file, "--golden", dir, "--sizes", "40x6,30x5"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	for _, name := range []string{"home_40x6.ansi", "home_30x5.ansi", "cart_40x6.ansi", "cart_30x5.ansi"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if name == "cart_40x6.ansi" && !strings.Contains(string(data), "CART") {
			t.Errorf("the cart drawing lacks its title:\n%s", data)
		}
	}
	stdout.Reset()
	if code := Run([]string{file, "--golden", dir, "--sizes", "40x6,30x5", "--check"}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "is current: 4 drawings") {
		t.Fatalf("check: exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(filepath.Join(dir, "cart_40x6.ansi"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := Run([]string{file, "--golden", dir, "--sizes", "40x6,30x5", "--check"}, &stdout, &stderr); code != 1 || !strings.Contains(stdout.String(), "cart_40x6.ansi: differs") {
		t.Fatalf("check of an edited drawing: exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if code := Run([]string{file, "--golden", dir, "--sizes", "wide"}, &stdout, &stderr); code != 1 {
		t.Fatalf("a bad size: exit %d", code)
	}
}

// With --generated the pages are drawn by the Go the export writes.
func TestGoldenDrawingsOfTheGeneratedGo(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	file := goldenDesign(t)
	dir := filepath.Join(t.TempDir(), "golden")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{file, "--golden", dir, "--generated"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, "cart_40x6.ansi"))
	if err != nil || !strings.Contains(string(data), "CART") {
		t.Fatalf("the generated cart: %v\n%s", err, data)
	}
	stdout.Reset()
	if code := Run([]string{file, "--golden", dir, "--generated", "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("check: exit %d: %s%s", code, stdout.String(), stderr.String())
	}
}
