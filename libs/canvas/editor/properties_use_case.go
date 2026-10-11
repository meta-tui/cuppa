package editor

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/meta-tui/cuppa/libs/catalog/definition"
	"github.com/meta-tui/cuppa/libs/color/space"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// Rename sets the node's display name.
func (e *Editor) Rename(id design.NodeID, name string) bool {
	name = strings.TrimSpace(name)
	n, ok := e.doc.Get(id)
	if !ok || name == "" || name == n.Name {
		return false
	}
	return e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Name = name })
	})
}

// SetProp sets one property after validating the value against the
// component's property specification.
func (e *Editor) SetProp(id design.NodeID, key, value string) error {
	n, ok := e.doc.Get(id)
	if !ok {
		return fmt.Errorf("editor: no node %q", id)
	}
	if n.Locked {
		return fmt.Errorf("editor: %s is locked", n.Name)
	}
	def, ok := e.cat.Get(n.Component)
	if !ok {
		return fmt.Errorf("editor: unknown component %q", n.Component)
	}
	spec, ok := def.Prop(key)
	if !ok {
		return fmt.Errorf("editor: %s has no property %q", def.ID, key)
	}
	var err error
	if ref, exists := e.isSwatchRef(value); ref && spec.Kind == definition.PropColor {
		if !exists {
			return fmt.Errorf("editor: there is no colour %s in the palette", strings.TrimSpace(value))
		}
		value = strings.TrimSpace(value)
	} else if value, err = normalise(spec, value); err != nil {
		return err
	}
	// A colour that follows the theme is only left alone when it already holds
	// this value: picking the colour it shows is still a choice to override.
	if cur, set := n.Props[key]; (set && cur == value) || (!set && spec.Default == value && spec.Role == definition.RoleNone) {
		return nil
	}
	e.apply(func() bool {
		if n.IsContainer() {
			// Pages and the shown page move the members with their bodies.
			return e.setContainerProps(n, map[string]string{key: value})
		}
		return e.doc.Update(id, func(n *design.Node) {
			if n.Props == nil {
				n.Props = map[string]string{}
			}
			n.Props[key] = value
		})
	})
	return nil
}

// normalise validates value for spec and returns its canonical form.
func normalise(spec definition.PropSpec, value string) (string, error) {
	switch spec.Kind {
	case definition.PropInt:
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return "", fmt.Errorf("editor: %s must be a whole number", spec.Label)
		}
		if spec.Max > spec.Min {
			n = min(max(n, spec.Min), spec.Max)
		}
		return strconv.Itoa(n), nil
	case definition.PropFloat:
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return "", fmt.Errorf("editor: %s must be a number", spec.Label)
		}
		if spec.Max > spec.Min {
			f = min(max(f, float64(spec.Min)), float64(spec.Max))
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	case definition.PropBool:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "yes", "on", "1":
			return "true", nil
		case "false", "no", "off", "0":
			return "false", nil
		}
		return "", fmt.Errorf("editor: %s must be true or false", spec.Label)
	case definition.PropColor:
		canon, err := space.Normalise(value)
		if err != nil {
			return "", fmt.Errorf("editor: %s: %w", spec.Label, err)
		}
		return canon, nil
	case definition.PropChoice:
		if !slices.Contains(spec.Choices, value) {
			return "", fmt.Errorf("editor: %s must be one of %s", spec.Label, strings.Join(spec.Choices, ", "))
		}
	}
	return value, nil
}

// ClearProp removes the component's own value for key, so the property follows
// the theme (or its default) again. It reports whether there was a value to
// remove.
func (e *Editor) ClearProp(id design.NodeID, key string) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Locked {
		return false
	}
	if _, set := n.Props[key]; !set {
		return false
	}
	return e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) {
			delete(n.Props, key)
			if len(n.Props) == 0 {
				n.Props = nil
			}
		})
	})
}
