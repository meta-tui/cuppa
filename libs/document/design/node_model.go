package design

// NodeID identifies one component instance inside a document.
type NodeID string

// Node is one placed component: which catalog entry it is, where it sits and
// the property values the designer changed.
type Node struct {
	ID        NodeID            `json:"id"`
	Component string            `json:"component"`
	Name      string            `json:"name"`
	Rect      Rect              `json:"rect"`
	// Layout holds size expressions ("50%", "100% - 10") for the axes of Rect
	// that follow the canvas; an empty one leaves that axis fixed. Rect is
	// always the resolved result at the current canvas size.
	Layout Layout `json:"layout,omitzero"`
	Props     map[string]string `json:"props,omitempty"`
	// Hidden nodes are not drawn, exported or hit by the pointer; they stay in
	// the layer list.
	Hidden bool `json:"hidden,omitempty"`
	// Locked nodes are drawn but cannot be moved, resized, deleted or have their
	// properties changed.
	Locked bool `json:"locked,omitempty"`
	// Draggable and Resizable are behaviours of the exported program, not of
	// the design: the person using the app can move the component (drag it
	// from anywhere) or resize it (drag its bottom-right corner).
	Draggable bool `json:"draggable,omitempty"`
	Resizable bool `json:"resizable,omitempty"`
	// Bind ties properties to named screen inputs: key is the property, value
	// the input's name. The value stored in Props stays as the default and as
	// what the design shows. See ADR 0007.
	Bind map[string]string `json:"bind,omitempty"`
	// ShowIf names a yes/no screen input; the component is drawn only while it
	// is true. Empty means always.
	ShowIf string `json:"showIf,omitempty"`
	// Event names the screen event raised when the component is clicked.
	Event string `json:"event,omitempty"`
	// Children are set on a group: the components it holds, with rectangles
	// relative to the group's top-left, laid out for a box of BaseW x BaseH. A
	// placed group scales them with its own size.
	Children []Node `json:"children,omitempty"`
	BaseW    int    `json:"baseW,omitempty"`
	BaseH    int    `json:"baseH,omitempty"`
	// In names the container (cuppa.tabs, cuppa.panels) this component is a
	// member of, and Page the page of it that shows the component; see ADR
	// 0011. A member is drawn only while its page is shown.
	In   NodeID `json:"in,omitempty"`
	Page int    `json:"page,omitempty"`
}

// The container components: their pages hold the components placed in them.
const (
	TabsComponent   = "cuppa.tabs"
	PanelsComponent = "cuppa.panels"
)

// IsContainer reports whether the node is a container of pages.
func (n Node) IsContainer() bool {
	return n.Component == TabsComponent || n.Component == PanelsComponent
}

// GroupComponent is the component id of a group.
const GroupComponent = "cuppa.group"

// IsGroup reports whether the node is a group of other nodes.
func (n Node) IsGroup() bool { return n.Component == GroupComponent && len(n.Children) > 0 }

// Clone returns a deep copy of the node.
func (n Node) Clone() Node {
	if n.Props != nil {
		props := make(map[string]string, len(n.Props))
		for k, v := range n.Props {
			props[k] = v
		}
		n.Props = props
	}
	if n.Children != nil {
		children := make([]Node, len(n.Children))
		for i, c := range n.Children {
			children[i] = c.Clone()
		}
		n.Children = children
	}
	return n
}

// Layout is the size expressions of a node, one per axis; see libs/layout/expr
// for the grammar. X and W are relative to the parent's width, Y and H to its
// height. An empty expression means that axis is fixed at its Rect value.
type Layout struct {
	X string `json:"x,omitempty"`
	Y string `json:"y,omitempty"`
	W string `json:"w,omitempty"`
	H string `json:"h,omitempty"`
}

// IsZero reports whether every axis is fixed.
func (l Layout) IsZero() bool { return l == Layout{} }
