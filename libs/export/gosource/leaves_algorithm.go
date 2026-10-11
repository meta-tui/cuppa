package gosource

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/container"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// Catalog is what the generator needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

// maxDepth stops a component that contains itself from expanding forever.
const maxDepth = 8

// leaf is one component the generated program places: a real component with
// its absolute position and final property values.
type leaf struct {
	// ID is the node the leaf came from (top level only), In the container it
	// is a member of and Page its page; Of is the index of the container's
	// leaf in the list, set by leaves.
	ID    design.NodeID
	In    design.NodeID
	Page  int
	Of    int
	Kind  string
	Name  string
	Rect  design.Rect
	Props map[string]string
	// Layout is set on a top-level component whose size or position follows
	// the window; Rect is then its size at the canvas size.
	Layout design.Layout
	// Drag and Resize are the behaviours the person using the program gets.
	Drag, Resize bool
	// MinW and MinH are the smallest size the component allows (set when it
	// follows the window or can be resized).
	MinW, MinH int
	// Bind, ShowIf and Event are the screen contract of the component (ADR
	// 0007): its bound properties (key to input name), the yes/no input that
	// shows it and the event a click raises.
	Bind          map[string]string
	ShowIf, Event string
	// Swatches maps the colour properties that use a palette swatch ("@Name")
	// to the swatch's name, so a screen can be handed another colour.
	Swatches map[string]string
	// Roles maps the colour properties the component has not set to the theme
	// role they follow, so a screen can be handed a theme at run time.
	Roles map[string]string
}

// leaves flattens a document back to front: groups are replaced by their
// parts, and components made from other components (pack components) by the
// parts they are made of, scaled to the size they were placed at.
func leaves(doc design.Document, cat Catalog) []leaf {
	t := themeOf{doc.Theme, doc.Background}
	// A member follows its container in the list, so it is drawn above it.
	ordered := make([]design.Node, 0, len(doc.Nodes))
	for _, i := range container.Order(doc) {
		ordered = append(ordered, doc.Nodes[i])
	}
	out := expand(ordered, design.Rect{W: doc.Width, H: doc.Height}, doc.Width, doc.Height, 0, nil, cat, t)
	at := map[design.NodeID]int{}
	for i, l := range out {
		if l.ID != "" {
			at[l.ID] = i
		}
	}
	for i := range out {
		if out[i].In == "" {
			continue
		}
		if c, ok := at[out[i].In]; ok && c < i {
			out[i].Of = c
		} else {
			out[i].In = ""
		}
	}
	return out
}

// themeOf is the design's theme and background, which colours that components
// do not set follow.
type themeOf struct {
	theme      design.Theme
	background string
}

// expand places nodes whose rectangles are relative to a box of baseW by baseH
// that sits at origin; override carries values a pack component hands down.
func expand(nodes []design.Node, origin design.Rect, baseW, baseH, depth int, override func(design.Node) map[string]string, cat Catalog, t themeOf) []leaf {
	var out []leaf
	for _, n := range nodes {
		if n.Hidden {
			continue
		}
		rect := n.Rect
		if origin.W != baseW || origin.H != baseH {
			rect = rect.Scale(baseW, baseH, origin.W, origin.H)
		}
		rect = rect.Translate(origin.X, origin.Y)
		props := withDefaults(n, cat, t)
		if override != nil {
			for k, v := range override(n) {
				props[k] = v
			}
		}
		switch {
		case n.Component == drawlayer.Component:
			// The drawing is only as many cells as were painted: one small
			// component per run of cells that share a style, so what is
			// under the empty cells shows through.
			out = append(out, drawRuns(n.Props[drawlayer.PropCells], rect)...)
		case n.IsGroup() && depth < maxDepth:
			out = append(out, expand(n.Children, rect, n.BaseW, n.BaseH, depth+1, nil, cat, t)...)
		case depth < maxDepth && isComposite(n, cat):
			def, _ := cat.Get(n.Component)
			inner := *def.Inner
			hand := func(part design.Node) map[string]string {
				vals := map[string]string{}
				for _, e := range inner.Props {
					if e.Target != part.ID {
						continue
					}
					if v, ok := props[e.Key]; ok {
						vals[e.TargetProp] = v
					} else {
						vals[e.TargetProp] = e.Default
					}
				}
				return vals
			}
			out = append(out, expand(inner.Nodes, rect, inner.W, inner.H, depth+1, hand, cat, t)...)
		default:
			event := n.Event
			if n.Goto != "" {
				// A click goes to a page: the router turns this event into the
				// change of page.
				event = gotoEvent(n.Goto)
			}
			l := leaf{Kind: n.Component, Name: n.Name, Rect: rect, Props: props,
				ID: n.ID, In: n.In, Page: n.Page,
				Bind: n.Bind, ShowIf: n.ShowIf, Event: event, Roles: rolesOf(n, cat), Swatches: swatchesOf(n, cat, t.theme)}
			if depth > 0 {
				l.ID, l.In = "", ""
			}
			if depth == 0 && (!n.Layout.IsZero() || n.Draggable || n.Resizable) {
				l.Layout = n.Layout
				l.Drag, l.Resize = n.Draggable, n.Resizable
				l.MinW, l.MinH = 1, 1
				if def, ok := cat.Get(n.Component); ok {
					l.MinW, l.MinH = max(def.MinSize.W, 1), max(def.MinSize.H, 1)
				}
			}
			out = append(out, l)
		}
	}
	return out
}

func isComposite(n design.Node, cat Catalog) bool {
	def, ok := cat.Get(n.Component)
	return ok && def.Inner != nil
}

// withDefaults is the node's properties over the catalog defaults.
func withDefaults(n design.Node, cat Catalog, t themeOf) map[string]string {
	if def, ok := cat.Get(n.Component); ok {
		return def.Effective(n.Props, t.theme, t.background)
	}
	props := map[string]string{}
	for k, v := range n.Props {
		props[k] = v
	}
	return props
}

// rolesOf lists the colour properties of the node that follow a theme role
// because the node does not set them.
func rolesOf(n design.Node, cat Catalog) map[string]string {
	def, ok := cat.Get(n.Component)
	if !ok {
		return nil
	}
	var roles map[string]string
	for _, spec := range def.Props {
		if spec.Role == definition.RoleNone {
			continue
		}
		if _, own := n.Props[spec.Key]; own {
			continue
		}
		if roles == nil {
			roles = map[string]string{}
		}
		roles[spec.Key] = string(spec.Role)
	}
	return roles
}

// swatchesOf lists the colour properties of the node that use a swatch of the
// palette.
func swatchesOf(n design.Node, cat Catalog, theme design.Theme) map[string]string {
	def, ok := cat.Get(n.Component)
	if !ok {
		return nil
	}
	var out map[string]string
	for _, spec := range def.Props {
		name, used := strings.CutPrefix(n.Props[spec.Key], design.TokenPrefix)
		if spec.Kind != definition.PropColor || !used {
			continue
		}
		if _, exists := theme.Colour(name); !exists {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		out[spec.Key] = name
	}
	return out
}
