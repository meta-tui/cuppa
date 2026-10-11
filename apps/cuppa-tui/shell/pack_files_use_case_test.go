package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/packsdialog"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func teaFile(t *testing.T, dir string) string {
	t.Helper()
	data, err := cupp.Encode(cupp.Pack{ID: "tea-shop", Name: "Tea shop", Components: []design.Composite{{
		ID: "card", Name: "Card", W: 20, H: 5,
		Nodes: []design.Node{{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "download.cupp")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func packsShell(t *testing.T) (*Model, string) {
	t.Helper()
	m := newShell(t)
	m.userPacks = filepath.Join(t.TempDir(), "packs")
	return m, m.userPacks
}

func TestInstallingAPackCopiesItAndAddsItsComponents(t *testing.T) {
	m, dir := packsShell(t)
	if err := m.installPack(teaFile(t, t.TempDir())); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tea-shop.cupp")); err != nil {
		t.Fatalf("the pack is copied under its own id: %v", err)
	}
	if _, ok := m.cat.Get("tea-shop.card"); !ok {
		t.Fatal("the component is in the catalog now")
	}
	m.pal.SetSize(m.layout.palette.W, 160)
	if !strings.Contains(paletteText(m), "Tea shop") {
		t.Fatalf("the palette lists the pack:\n%s", paletteText(m))
	}
	if err := m.installPack(teaFile(t, t.TempDir())); err == nil {
		t.Fatal("the same pack twice is refused")
	}
}

func TestAddingSomethingThatIsNotAPackIsRefused(t *testing.T) {
	m, dir := packsShell(t)
	bad := filepath.Join(t.TempDir(), "bad.cupp")
	if err := os.WriteFile(bad, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := m.installPack(bad); err == nil {
		t.Fatal("a damaged file is refused")
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("nothing is created for a refused pack")
	}
}

func TestRemovingAPackDeletesItsFileAfterAsking(t *testing.T) {
	m, dir := packsShell(t)
	if err := m.installPack(teaFile(t, t.TempDir())); err != nil {
		t.Fatal(err)
	}
	m.afterPacksDialog(modal.Outcome{Button: packsdialog.RemoveButton, Value: "tea-shop"})
	dlg := m.flow.Modal()
	if dlg == nil {
		t.Fatal("removing asks first")
	}
	if _, err := os.Stat(filepath.Join(dir, "tea-shop.cupp")); err != nil {
		t.Fatal("nothing is deleted before the answer")
	}
	dlg.Key("", false, true, false) // Enter: the first button, Remove
	m.flow.Resolve()
	if _, err := os.Stat(filepath.Join(dir, "tea-shop.cupp")); err == nil {
		t.Fatal("the file is deleted")
	}
	if _, ok := m.cat.Get("tea-shop.card"); ok {
		t.Fatal("the pack is gone from the catalog")
	}
}

func TestBundledPacksAreInTheCatalogAndCannotBeRemoved(t *testing.T) {
	m := newShell(t)
	m.reloadPacks() // what LoadUserPacks does, without reading the user's folder
	if _, ok := m.cat.Get("starter.card"); !ok {
		t.Fatal("the starter pack ships with Cuppa")
	}
	for _, p := range m.cat.Packs() {
		if p.ID == "starter" && (!p.Builtin || p.Source != "") {
			t.Fatalf("a bundled pack is built in, with no file to delete: %+v", p)
		}
	}
	m.afterPacksDialog(modal.Outcome{Button: packsdialog.RemoveButton, Value: "starter"})
	if m.flow.Modal() != nil {
		t.Fatal("removing a bundled pack asks nothing and does nothing")
	}
}
