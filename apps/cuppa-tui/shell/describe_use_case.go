package shell

import (
	"fmt"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
)

// Describe is what a screen reader should hear about the screen right now. An
// open dialog is the whole screen, as it is for the pointer; otherwise the
// menus, the palette, the canvas and the details bar, in that order.
func (m *Model) Describe() a11y.Snapshot {
	snap := a11y.Snapshot{Title: "Cuppa, " + m.flow.Title(), Status: m.flow.Status()}
	if m.announce != "" {
		snap.Status = m.announce
	}
	if m.ed.Dirty() {
		snap.Title += ", unsaved changes"
	}
	if dlg := m.flow.Modal(); dlg != nil {
		if d, ok := dlg.(a11y.Describer); ok {
			snap.Nodes = d.Describe()
		} else {
			snap.Nodes = []a11y.Node{a11y.Heading("Dialog")}
		}
		return snap
	}
	doc := m.ed.Document()
	if m.run != nil {
		snap.Nodes = append(snap.Nodes, a11y.Text("Preview running. Escape goes back to editing."))
	}
	snap.Nodes = append(snap.Nodes, m.focused(inMenu, m.bar.Describe())...)
	snap.Nodes = append(snap.Nodes, m.focused(inPalette, m.pal.Describe())...)
	snap.Nodes = append(snap.Nodes, m.focused(inStage, []a11y.Node{
		a11y.Heading("Canvas"),
		a11y.Text(fmt.Sprintf("%d by %d cells, %d components", doc.Width, doc.Height, len(doc.Nodes))),
	})...)
	snap.Nodes = append(snap.Nodes, m.focused(inInspector, m.ins.Describe())...)
	return snap
}

// focused flags the first node of an area when the keyboard is on it, so a
// screen reader announces where it is.
func (m *Model) focused(area pane, nodes []a11y.Node) []a11y.Node {
	if m.focus == area && len(nodes) > 0 {
		nodes[0].Focused = true
	}
	return nodes
}
