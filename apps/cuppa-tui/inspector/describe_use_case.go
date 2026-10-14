package inspector

import (
	"fmt"
	"strconv"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Describe says what the details bar shows: the selection's settings, or the
// canvas options when nothing is selected, then the layers.
func (m *Model) Describe() []a11y.Node {
	nodes := []a11y.Node{a11y.Heading("Details")}
	nodes = append(nodes, a11y.List("History", onOff("Undo", m.ed.CanUndo()), onOff("Redo", m.ed.CanRedo())))
	selected := m.ed.Selected()
	switch n, ok := m.ed.Primary(); {
	case !ok:
		nodes = append(nodes, m.describeCanvas()...)
	case len(selected) > 1:
		nodes = append(nodes, a11y.Text(fmt.Sprintf("%d components selected", len(selected))),
			onOff("Group", m.ed.CanGroup()), onOff("Ungroup", m.ed.CanUngroup()))
	default:
		nodes = append(nodes, m.describeNode(n)...)
	}
	if m.message != "" {
		nodes = append(nodes, a11y.Text(m.message))
	}
	return append(nodes, m.describeLayers())
}

// onOff is a button that may not be usable right now.
func onOff(label string, usable bool) a11y.Node {
	if !usable {
		label += ", not available"
	}
	return a11y.Button(label)
}

func (m *Model) describeCanvas() []a11y.Node {
	doc := m.ed.Document()
	bg := doc.Background
	if bg == "" {
		bg = "terminal default"
	}
	nodes := []a11y.Node{
		a11y.Heading("Canvas"),
		a11y.Text("Nothing selected. Drag a component from the palette onto the canvas."),
		a11y.Field("Width", strconv.Itoa(doc.Width)),
		a11y.Field("Height", strconv.Itoa(doc.Height)),
		a11y.Field("Background", bg),
	}
	for _, e := range []struct{ label, name string }{
		{"Grid dots", editor.EffectGrid}, {"Shadows", editor.EffectShadow},
		{"Scanlines", editor.EffectScanlines}, {"Vignette", editor.EffectVignette},
	} {
		b := a11y.Button(e.label)
		b.Value = state(m.ed.Effect(e.name))
		nodes = append(nodes, b)
	}
	terminal := a11y.Button("Preview on a light terminal")
	terminal.Value = state(doc.Light)
	profile := "true colour"
	for _, p := range profiles {
		if p.value == doc.Profile {
			profile = p.name
		}
	}
	nodes = append(nodes, terminal, a11y.Field("Colours", profile))
	if m.snapGet != nil {
		b := a11y.Button("Snap to guides")
		b.Value = state(m.snapGet())
		nodes = append(nodes, b)
	}
	return nodes
}

func state(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

func (m *Model) describeNode(n design.Node) []a11y.Node {
	doc := m.ed.Document()
	nodes := []a11y.Node{
		a11y.Heading(n.Name),
		a11y.Field("Name", n.Name),
		a11y.Text("Type " + n.Component),
	}
	if n.Locked {
		nodes = append(nodes, a11y.Text("Locked: unlock it in Layers to edit"))
	}
	nodes = append(nodes,
		a11y.Field("X", strconv.Itoa(n.Rect.X)), a11y.Field("Y", strconv.Itoa(n.Rect.Y)),
		a11y.Field("Width", strconv.Itoa(n.Rect.W)), a11y.Field("Height", strconv.Itoa(n.Rect.H)),
		a11y.Text(fmt.Sprintf("Layer %d of %d", doc.Index(n.ID)+1, len(doc.Nodes))),
		a11y.Button("Bring to front"), a11y.Button("Bring forward"), a11y.Button("Send backward"), a11y.Button("Send to back"),
		a11y.Button("Duplicate"), a11y.Button("Delete"),
		onOff("Group", m.ed.CanGroup()), onOff("Ungroup", m.ed.CanUngroup()),
	)
	if def, ok := m.cat.Get(n.Component); ok {
		var props []a11y.Node
		for _, p := range def.Props {
			value, set := n.Props[p.Key]
			if !set {
				value = p.Default
			}
			if p.Kind == definition.PropMenu {
				value = describeMenus(value)
			}
			if value == "" {
				value = "empty"
			}
			props = append(props, a11y.Field(p.Label, value))
		}
		if len(props) > 0 {
			nodes = append(nodes, a11y.List("Properties", props...))
		}
	}
	return nodes
}

// describeLayers lists the components front to back with what the eye and
// padlock icons say.
func (m *Model) describeLayers() a11y.Node {
	doc := m.ed.Document()
	var items []a11y.Node
	for i := len(doc.Nodes) - 1; i >= 0; i-- {
		n := doc.Nodes[i]
		label := fmt.Sprintf("%s, %s, at %d,%d, %d by %d", n.Name, n.Component, n.Rect.X, n.Rect.Y, n.Rect.W, n.Rect.H)
		if n.Hidden {
			label += ", hidden"
		}
		if n.Locked {
			label += ", locked"
		}
		items = append(items, a11y.Item(label, m.ed.IsSelected(n.ID)))
	}
	if len(items) == 0 {
		return a11y.Text("Layers: none")
	}
	return a11y.List("Layers, front to back", items...)
}
