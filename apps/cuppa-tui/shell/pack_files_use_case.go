package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/confirm"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/filedialog"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/packsdialog"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

// afterPacksDialog acts on how the Packs dialog ended: Add and Remove change
// the packs folder and then show the dialog again, so the list is current.
func (m *Model) afterPacksDialog(o modal.Outcome) {
	switch o.Button {
	case packsdialog.AddButton:
		m.addPack()
	case packsdialog.RemoveButton:
		m.removePack(definition.Family(o.Value))
	}
}

// addPack asks for a .cupp file and installs a copy of it in the packs folder.
func (m *Model) addPack() {
	if m.userPacks == "" {
		m.flow.Notice("Cannot add a pack", "Cuppa has no folder for packs on this computer.")
		return
	}
	d := filedialog.New(filedialog.Spec{Title: "Add component pack", Ext: cupp.Extension})
	m.flow.Show(d, func(o modal.Outcome) {
		if o.Canceled {
			m.openPacks()
			return
		}
		if err := m.installPack(o.Path); err != nil {
			m.flow.Notice("Cannot add this pack", err.Error())
			return
		}
		m.openPacks()
	})
}

// installPack validates the file, refuses an id that is taken, and copies it
// to <packs folder>/<pack id>.cupp.
func (m *Model) installPack(path string) error {
	p, _, err := cupp.LoadFile(path)
	if err != nil {
		return err
	}
	for _, have := range m.cat.Packs() {
		if string(have.ID) == p.ID {
			return fmt.Errorf("a pack with the id %q is already there; remove it first", p.ID)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(m.userPacks, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(m.userPacks, p.ID+cupp.Extension), data, 0o644); err != nil {
		return err
	}
	m.reloadPacks()
	return nil
}

// removePack deletes an installed pack's file after asking.
func (m *Model) removePack(id definition.Family) {
	var pack definition.Pack
	for _, p := range m.cat.Packs() {
		if p.ID == id {
			pack = p
		}
	}
	if pack.Source == "" {
		return
	}
	body := fmt.Sprintf("Remove the pack %q?\nThis deletes %s.\nDesigns that use it keep opening, with placeholders.", pack.Name, filepath.Base(pack.Source))
	m.flow.Show(confirm.New("Remove pack", body, "Remove", "Cancel"), func(o modal.Outcome) {
		if o.Canceled || o.Button != "Remove" {
			m.openPacks()
			return
		}
		if err := os.Remove(pack.Source); err != nil {
			m.flow.Notice("Cannot remove the pack", err.Error())
			return
		}
		m.reloadPacks()
		m.openPacks()
	})
}
