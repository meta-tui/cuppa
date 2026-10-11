package shell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Where components made by the user are kept.
const (
	personalPackID   = "my-components"
	personalPackName = "My components"
)

// canSaveComponent reports whether the selection can become a component: a
// group, or components that can be grouped first.
func (m *Model) canSaveComponent() bool {
	if n, ok := m.ed.Primary(); ok && len(m.ed.Selected()) == 1 && n.IsGroup() {
		return !n.Locked
	}
	return m.ed.CanGroup()
}

// saveAsComponent turns the selection into a component in the personal pack.
// Components that are not a group yet are grouped first (one undo step).
func (m *Model) saveAsComponent() {
	if m.userPacks == "" {
		m.flow.Notice("Cannot save a component", "Cuppa has no folder for packs on this computer.")
		return
	}
	if !m.canSaveComponent() {
		m.flow.Notice("Save as component", "Select a group, or two or more components,\nthat are not locked.")
		return
	}
	n, _ := m.ed.Primary()
	if len(m.ed.Selected()) > 1 || !n.IsGroup() {
		m.ed.Group()
		n, _ = m.ed.Primary()
	}
	group := n
	m.flow.Show(prompt.New("Save as component", "Name of the component", group.Name), func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		name, err := m.storeComponent(group, o.Value)
		if err != nil {
			m.flow.Notice("Cannot save the component", err.Error())
			return
		}
		m.flow.Notice("Component saved", fmt.Sprintf("%q is in the pack %q.\nFind it in the palette.", name, personalPackName))
	})
}

// storeComponent adds the group to the personal pack file and reloads the packs.
func (m *Model) storeComponent(group design.Node, name string) (string, error) {
	path := filepath.Join(m.userPacks, personalPackID+cupp.Extension)
	pack := cupp.Pack{ID: personalPackID, Name: personalPackName, Description: "Components made in Cuppa"}
	if _, err := os.Stat(path); err == nil {
		loaded, _, err := cupp.LoadFile(path)
		if err != nil {
			return "", err
		}
		pack = loaded
	}
	c := cupp.FromGroup(group, name, m.cat)
	cupp.AddComponent(&pack, c)
	data, err := cupp.Encode(pack)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(m.userPacks, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	m.reloadPacks()
	return name, nil
}
