package editor

import (
	"fmt"
	"slices"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/compat"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// lister is the catalog that can say which components it has.
type lister interface {
	List() []definition.Definition
}

// ChangeOptions lists the components a component can be changed for without
// losing the variables bound to it, in catalog order: the ones that carry every
// port its bindings use (a text input and a colour picker both carry a value, a
// tab bar and dialog buttons both carry items and the chosen one). A group, a
// pack component and a locked component have none.
func (e *Editor) ChangeOptions(id design.NodeID) []definition.Definition {
	n, from, ok := e.changeable(id)
	if !ok || n.Locked {
		return nil
	}
	l, ok := e.cat.(lister)
	if !ok {
		return nil
	}
	return compat.Candidates(l.List(), from, n.BoundKeys())
}

// ChangeReport says what changing the component for another would keep, lose
// and add, without doing it.
func (e *Editor) ChangeReport(id design.NodeID, component string) (compat.Report, error) {
	n, from, ok := e.changeable(id)
	if !ok {
		return compat.Report{}, fmt.Errorf("editor: %q is not a component that can be changed", id)
	}
	to, ok := e.cat.Get(component)
	if !ok {
		return compat.Report{}, fmt.Errorf("editor: unknown component %q", component)
	}
	return compat.Check(from, to, n.BoundKeys()), nil
}

// ChangeComponent changes a component for another one in place: it keeps its
// name, position, size (made larger if the new component needs more), layout,
// show-if and event; the properties that carry the same port keep their
// bindings and, when the value is valid for the new property, their values;
// other properties with the same key and kind keep their values too. It is one
// undo step.
//
// A binding that the new component has no property for is lost; unless
// allowLoss is true that is an error, and nothing changes.
func (e *Editor) ChangeComponent(id design.NodeID, component string, allowLoss bool) (compat.Report, error) {
	n, from, ok := e.changeable(id)
	if !ok {
		return compat.Report{}, fmt.Errorf("editor: %q is not a component that can be changed", id)
	}
	if n.Locked {
		return compat.Report{}, fmt.Errorf("editor: %s is locked", n.Name)
	}
	to, ok := e.cat.Get(component)
	if !ok {
		return compat.Report{}, fmt.Errorf("editor: unknown component %q", component)
	}
	if to.Inner != nil {
		return compat.Report{}, fmt.Errorf("editor: %s is a pack component; a component is changed for a built-in one", to.Name)
	}
	if to.ID == from.ID {
		return compat.Report{}, nil
	}
	report := compat.Check(from, to, n.BoundKeys())
	if !report.Compatible() && !allowLoss {
		var names []string
		for _, key := range report.Lost {
			label := key
			if spec, ok := from.Prop(key); ok {
				label = spec.Label
			}
			names = append(names, fmt.Sprintf("%s (%s)", label, n.Bind[key]))
		}
		return report, fmt.Errorf("editor: %s has nothing for %s; changing would remove those variables", to.Name, strings.Join(names, ", "))
	}

	props := map[string]string{}
	taken := map[string]bool{}
	carry := func(fromKey, toKey string) {
		value, set := n.Props[fromKey]
		spec, _ := to.Prop(toKey)
		if !set {
			return
		}
		if ref, exists := e.isSwatchRef(value); ref {
			if exists && spec.Kind == definition.PropColor {
				props[toKey] = strings.TrimSpace(value)
			}
			return
		}
		if v, err := normalise(spec, value); err == nil {
			props[toKey] = v
		}
	}
	bind := map[string]string{}
	for _, m := range report.Moves {
		taken[m.To] = true
		carry(m.From, m.To)
		if name := n.Bind[m.From]; name != "" {
			bind[m.To] = name
		}
	}
	for _, p := range from.Props {
		spec, ok := to.Prop(p.Key)
		if !ok || taken[p.Key] || spec.Kind != p.Kind || p.Port != definition.PortNone || spec.Port != definition.PortNone {
			continue
		}
		if spec.Kind == definition.PropChoice && !slices.Equal(spec.Choices, p.Choices) {
			continue
		}
		carry(p.Key, p.Key)
	}

	rect := n.Rect
	rect.W, rect.H = max(rect.W, to.MinSize.W), max(rect.H, to.MinSize.H)
	e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) {
			n.Component = to.ID
			n.Props = nil
			if len(props) > 0 {
				n.Props = props
			}
			n.Bind = nil
			if len(bind) > 0 {
				n.Bind = bind
			}
			n.Rect = rect
		})
	})
	return report, nil
}

// changeable returns the node and its definition when it is a plain component:
// not a group, not a pack component, not the drawing layer.
func (e *Editor) changeable(id design.NodeID) (design.Node, definition.Definition, bool) {
	n, ok := e.doc.Get(id)
	if !ok || n.IsGroup() {
		return n, definition.Definition{}, false
	}
	def, ok := e.cat.Get(n.Component)
	if !ok || def.Inner != nil {
		return n, def, false
	}
	return n, def, true
}
