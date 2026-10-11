package shell

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

func painted(m *Model) string {
	g := scene.Render(m.ed.Document(), m.cat)
	var out strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if ch := g.At(x, y).Ch; ch != 0 {
				out.WriteRune(ch)
			}
		}
	}
	return out.String()
}

func TestADesignCarriesItsCustomComponentsToAComputerWithoutThePack(t *testing.T) {
	// Computer A has the pack and saves a design that uses it.
	a, _ := packsShell(t)
	if err := a.installPack(teaFile(t, t.TempDir())); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ed.Add("tea-shop.card", 2, 2); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "shop.cuppa")
	saved, err := disk.Save(file, cupp.Embed(a.ed.Document(), a.cat))
	if err != nil {
		t.Fatal(err)
	}

	// Computer B has no packs at all.
	b := newShell(t)
	if _, ok := b.cat.Get("tea-shop.card"); ok {
		t.Fatal("B must not have the pack")
	}
	if err := b.OpenFile(saved); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.cat.Get("tea-shop.card"); !ok {
		t.Fatal("opening the design adopts the embedded component")
	}
	if !strings.ContainsAny(painted(b), "╭┌┏╔") {
		t.Fatalf("the component draws, not a placeholder:\n%s", painted(b))
	}

	// Opening something else drops what it carried.
	b.flow.NewDesign()
	b.flow.Resolve()
	if _, ok := b.cat.Get("tea-shop.card"); ok {
		t.Fatal("a new design has no embedded components")
	}
}
