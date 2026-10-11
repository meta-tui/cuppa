package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// maxCompositeDepth stops a component that contains itself, directly or
// through others, from recursing forever.
const maxCompositeDepth = 8

// paintComposite paints the parts of a user-made component onto g. Parts are
// scaled from the component's default size to the size it was placed at, and
// the component's own property values are handed down to the parts they drive.
func paintComposite(g *grid.Grid, c design.Composite, props Props, cat Catalog, depth int, theme design.Theme, background string) {
	if depth >= maxCompositeDepth {
		paintGeneric(g, c.Name)
		return
	}
	for _, part := range c.Nodes {
		if part.Hidden {
			continue
		}
		part = part.Clone()
		if part.Props == nil {
			part.Props = map[string]string{}
		}
		for _, e := range c.Props {
			if e.Target != part.ID {
				continue
			}
			if v, ok := props[e.Key]; ok {
				part.Props[e.TargetProp] = v
			} else {
				part.Props[e.TargetProp] = e.Default
			}
		}
		part.Rect = part.Rect.Scale(c.W, c.H, g.W, g.H)
		g.Blit(renderNode(part, cat, depth+1, theme, background), part.Rect.X, part.Rect.Y)
	}
}
