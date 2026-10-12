package editor

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// SetSwatch adds a named colour to the design's palette, or changes the colour
// of the one that has the name. Properties set to "@Name" follow it. One undo
// step.
func (e *Editor) SetSwatch(name, color string) error {
	name = design.CleanSwatchName(name)
	if !design.ValidSwatchName(name) {
		return fmt.Errorf("editor: a colour name starts with a letter and has letters, digits, spaces, - or _")
	}
	color, err := space.Normalise(color)
	if err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	if color == "" {
		return fmt.Errorf("editor: a palette colour needs a value")
	}
	if cur, ok := e.doc.Theme.Colour(name); ok && cur == color {
		return nil
	}
	e.apply(func() bool {
		palette := append([]design.Swatch(nil), e.doc.Theme.Palette...)
		for i := range palette {
			if palette[i].Name == name {
				palette[i].Color = color
				e.doc.Theme.Palette = palette
				return true
			}
		}
		e.doc.Theme.Palette = append(palette, design.Swatch{Name: name, Color: color})
		return true
	})
	return nil
}

// RemoveSwatch deletes a named colour. It refuses while a property uses it, so
// no colour silently becomes empty.
func (e *Editor) RemoveSwatch(name string) error {
	if _, ok := e.doc.Theme.Colour(name); !ok {
		return nil
	}
	if n := e.swatchUses(name); n > 0 {
		return fmt.Errorf("editor: %s is used by %d properties; set them to another colour first", name, n)
	}
	e.apply(func() bool {
		var palette []design.Swatch
		for _, s := range e.doc.Theme.Palette {
			if s.Name != name {
				palette = append(palette, s)
			}
		}
		e.doc.Theme.Palette = palette
		return true
	})
	return nil
}

// swatchUses counts the properties set to the swatch.
func (e *Editor) swatchUses(name string) int {
	ref := design.TokenPrefix + name
	count := 0
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for _, n := range nodes {
			def, known := e.cat.Get(n.Component)
			for k, v := range n.Props {
				if v != ref {
					continue
				}
				if spec, ok := def.Prop(k); known && ok && spec.Kind == definition.PropColor {
					count++
				}
			}
			walk(n.Children)
		}
	}
	walk(e.doc.Nodes)
	return count
}

// isSwatchRef reports whether value is a reference to a palette swatch, and
// whether that swatch exists.
func (e *Editor) isSwatchRef(value string) (ref, exists bool) {
	name, ok := strings.CutPrefix(strings.TrimSpace(value), design.TokenPrefix)
	if !ok {
		return false, false
	}
	_, exists = e.doc.Theme.Colour(name)
	return true, exists
}

// SetColour sets a colour property the way the details bar does: the colour
// belongs to a named palette swatch, so other components can use the same name
// and follow it, and a program that shows the screen can change it.
//
// When the property already uses a swatch, that swatch changes (and so does
// every other property that uses it). Otherwise a new swatch is made, named
// after the property ("Foreground", then "Foreground 2"...), and the property
// uses it. An empty value gives the property no colour of its own and leaves
// the palette alone. One undo step.
func (e *Editor) SetColour(id design.NodeID, key, value string) error {
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
	if !ok || spec.Kind != definition.PropColor {
		return fmt.Errorf("editor: %s has no colour property %q", def.ID, key)
	}
	color, err := space.Normalise(value)
	if err != nil {
		return fmt.Errorf("editor: %s: %w", spec.Label, err)
	}
	if color == "" {
		return e.SetProp(id, key, "")
	}
	if name, used := strings.CutPrefix(n.Props[key], design.TokenPrefix); used {
		if _, exists := e.doc.Theme.Colour(name); exists {
			return e.SetSwatch(name, color)
		}
	}
	name := e.freshSwatchName(spec.Label)
	e.apply(func() bool {
		e.doc.Theme.Palette = append(append([]design.Swatch(nil), e.doc.Theme.Palette...), design.Swatch{Name: name, Color: color})
		return e.doc.Update(id, func(n *design.Node) {
			if n.Props == nil {
				n.Props = map[string]string{}
			}
			n.Props[key] = design.TokenPrefix + name
		})
	})
	return nil
}

// UseSwatch makes a colour property follow the named swatch, so the two stay
// in step. One undo step.
func (e *Editor) UseSwatch(id design.NodeID, key, name string) error {
	return e.SetProp(id, key, design.TokenPrefix+name)
}

// Swatches lists the palette, in the order the colours were named.
func (e *Editor) Swatches() []design.Swatch {
	return append([]design.Swatch(nil), e.doc.Theme.Palette...)
}

// freshSwatchName is a palette name made from a property's label that no
// swatch has.
func (e *Editor) freshSwatchName(label string) string {
	var b strings.Builder
	for _, r := range label {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	base := design.CleanSwatchName(b.String())
	if !design.ValidSwatchName(base) {
		base = "Colour"
	}
	name := base
	for i := 2; ; i++ {
		if _, taken := e.doc.Theme.Colour(name); !taken {
			return name
		}
		name = fmt.Sprintf("%s %d", base, i)
	}
}

// RenameSwatch gives a palette colour another name everywhere it is used, as
// one undo step. It refuses a name another colour already has.
func (e *Editor) RenameSwatch(from, to string) error {
	to = design.CleanSwatchName(to)
	if !design.ValidSwatchName(to) {
		return fmt.Errorf("editor: a name starts with a letter and has letters, digits, spaces, - or _")
	}
	if from == to {
		return nil
	}
	if _, ok := e.doc.Theme.Colour(from); !ok {
		return fmt.Errorf("editor: no colour called %s", from)
	}
	if _, ok := e.doc.Theme.Colour(to); ok {
		return fmt.Errorf("editor: there is already a colour called %s", to)
	}
	e.apply(func() bool {
		palette := append([]design.Swatch(nil), e.doc.Theme.Palette...)
		for i := range palette {
			if palette[i].Name == from {
				palette[i].Name = to
			}
		}
		e.doc.Theme.Palette = palette
		renameSwatchIn(e.doc.Nodes, from, to)
		for i := range e.doc.Pages {
			renameSwatchIn(e.doc.Pages[i].Nodes, from, to)
		}
		return true
	})
	return nil
}

// renameSwatchIn rewrites every use of a palette colour in the nodes, groups
// included.
func renameSwatchIn(nodes []design.Node, from, to string) {
	for i := range nodes {
		n := &nodes[i]
		for k, v := range n.Props {
			if v == design.TokenPrefix+from {
				n.Props[k] = design.TokenPrefix + to
			}
		}
		renameSwatchIn(n.Children, from, to)
	}
}
