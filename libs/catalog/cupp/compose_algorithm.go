package cupp

import (
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Catalog is what FromGroup needs to find the properties of the grouped parts.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

// maxExposed caps how many properties a new component offers, so a big group
// does not bury the details bar.
const maxExposed = 12

// Slug turns a display name into an id: lowercase letters and digits joined by
// single dashes. A name with none gives "component".
func Slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			b.WriteRune(r)
		} else {
			dash = true
		}
	}
	if b.Len() == 0 {
		return "component"
	}
	if s := b.String(); len(s) > 48 {
		return strings.TrimRight(s[:48], "-")
	}
	return b.String()
}

// FromGroup turns a group into a component named name, at the group's current
// size. The text properties of its parts (titles, labels, items…) become the
// component's own properties, so each placed copy can be edited.
func FromGroup(g design.Node, name string, cat Catalog) design.Composite {
	c := design.Composite{
		ID: Slug(name), Name: name, W: g.Rect.W, H: g.Rect.H,
	}
	for _, part := range g.Children {
		part = part.Clone()
		part.Rect = part.Rect.Scale(g.BaseW, g.BaseH, g.Rect.W, g.Rect.H)
		c.Nodes = append(c.Nodes, part)
	}
	keys := map[string]bool{}
	for _, part := range c.Nodes {
		def, ok := cat.Get(part.Component)
		if !ok {
			continue
		}
		for _, p := range def.Props {
			if p.Kind != definition.PropText || len(c.Props) >= maxExposed {
				continue
			}
			key := Slug(part.Name) + "-" + p.Key
			for n := 2; keys[key]; n++ {
				key = fmt.Sprintf("%s-%d", Slug(part.Name)+"-"+p.Key, n)
			}
			keys[key] = true
			value, set := part.Props[p.Key]
			if !set {
				value = p.Default
			}
			c.Props = append(c.Props, design.Exposed{
				Key: key, Label: part.Name + " " + strings.ToLower(p.Label), Kind: string(definition.PropText),
				Default: value, Target: part.ID, TargetProp: p.Key,
			})
		}
	}
	return c
}

// AddComponent puts c in the pack under a free id (the slug, then -2, -3…) and
// returns the id it got.
func AddComponent(p *Pack, c design.Composite) string {
	base := c.ID
	for n := 2; ; n++ {
		taken := false
		for _, have := range p.Components {
			if have.ID == c.ID {
				taken = true
			}
		}
		if !taken {
			break
		}
		c.ID = fmt.Sprintf("%s-%d", base, n)
	}
	p.Components = append(p.Components, c)
	return c.ID
}
