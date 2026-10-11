// Package container is what tabs and panels do: which page of a container is
// shown, where each page's body is, which member components are visible and
// what a press on a tab or a header changes. It is pure, so the designer, the
// exported program and the screens all draw and answer presses the same way
// (ADR 0011).
package container

import (
	"strconv"
	"strings"

	"github.com/meta-tui/cuppa/libs/document/design"
)

// Props of a container.
const (
	PropPages     = "pages"     // page names, comma separated
	PropActive    = "active"    // tabs: the page shown, 0-based
	PropOpen      = "open"      // panels: the open sections, comma separated
	PropExclusive = "exclusive" // panels: opening one closes the others
)

// DefaultPages is what a container starts with.
const DefaultPages = "Page 1,Page 2"

// TabBarRows is the height of the tab bar: the names and a rule under them.
const TabBarRows = 2

// Pages are the page names of a container, never empty.
func Pages(n design.Node) []string {
	var out []string
	for _, p := range strings.Split(n.Props[PropPages], ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"Page 1"}
	}
	return out
}

// Active is the page a tabs container shows, kept inside its pages.
func Active(n design.Node) int {
	i, _ := strconv.Atoi(strings.TrimSpace(n.Props[PropActive]))
	return min(max(i, 0), len(Pages(n))-1)
}

// OpenSet is the sections of a panels container that are open.
func OpenSet(n design.Node) map[int]bool {
	out := map[int]bool{}
	count := len(Pages(n))
	for _, f := range strings.Split(n.Props[PropOpen], ",") {
		if i, err := strconv.Atoi(strings.TrimSpace(f)); err == nil && i >= 0 && i < count {
			out[i] = true
		}
	}
	return out
}

// Exclusive reports whether opening a section closes the others.
func Exclusive(n design.Node) bool { return n.Props[PropExclusive] == "true" }

// Shown reports whether a page of the container is shown: the active tab, or
// an open section.
func Shown(n design.Node, page int) bool {
	if n.Component == design.TabsComponent {
		return page == Active(n)
	}
	return OpenSet(n)[page]
}

// Tab is one tab of the bar: its page and the columns it takes.
type Tab struct {
	Page   int
	X0, X1 int
}

// Head is the header row of one section.
type Head struct {
	Page int
	Y    int
}

// Geometry is where the parts of a container are, in canvas cells.
type Geometry struct {
	// Tabs and Rule are the bar of a tabs container; Rule is its row.
	Tabs []Tab
	Rule int
	// Heads are the headers of a panels container.
	Heads []Head
	// Bodies is the rectangle of each page's body. A closed section's body has
	// no height; its top is where its members keep their offset from.
	Bodies []design.Rect
}

// Layout computes the geometry of a container at its current rectangle and
// state. Tabs: the names on the first row (each as " name ", a bar between
// them), a rule under them, and every page's body is the rest. Panels: one
// header row per section, and the rows left after the headers are shared by
// the open sections, the last open one taking the remainder.
func Layout(n design.Node) Geometry {
	pages := Pages(n)
	r := n.Rect
	var g Geometry
	if n.Component == design.TabsComponent {
		x := r.X
		for i, name := range pages {
			w := len([]rune(name)) + 2
			g.Tabs = append(g.Tabs, Tab{Page: i, X0: x, X1: min(x+w, r.X+r.W)})
			x += w + 1
		}
		g.Rule = r.Y + 1
		body := design.Rect{X: r.X, Y: r.Y + TabBarRows, W: r.W, H: max(r.H-TabBarRows, 0)}
		g.Bodies = make([]design.Rect, len(pages))
		for i := range g.Bodies {
			g.Bodies[i] = body
		}
		return g
	}
	open := OpenSet(n)
	openCount := len(open)
	avail := max(r.H-len(pages), 0)
	y := r.Y
	g.Bodies = make([]design.Rect, len(pages))
	left, opened := avail, 0
	for i := range pages {
		g.Heads = append(g.Heads, Head{Page: i, Y: y})
		y++
		h := 0
		if open[i] && openCount > 0 {
			opened++
			h = avail / openCount
			if opened == openCount {
				h = left
			}
			left -= h
		}
		g.Bodies[i] = design.Rect{X: r.X, Y: y, W: r.W, H: h}
		y += h
	}
	return g
}

// HeaderAt reports which tab or section header is at the cell, if any.
func HeaderAt(n design.Node, x, y int) (int, bool) {
	g := Layout(n)
	if n.Component == design.TabsComponent {
		if y != n.Rect.Y {
			return 0, false
		}
		for _, t := range g.Tabs {
			if x >= t.X0 && x < t.X1 {
				return t.Page, true
			}
		}
		return 0, false
	}
	if x < n.Rect.X || x >= n.Rect.X+n.Rect.W {
		return 0, false
	}
	for _, h := range g.Heads {
		if h.Y == y {
			return h.Page, true
		}
	}
	return 0, false
}

// Pressed is the properties of the container after a press on page's tab or
// header: a tab becomes the active one, a section flips (and in an accordion
// closes the others).
func Pressed(n design.Node, page int) map[string]string {
	if n.Component == design.TabsComponent {
		return map[string]string{PropActive: strconv.Itoa(page)}
	}
	open := OpenSet(n)
	if open[page] {
		delete(open, page)
	} else {
		if Exclusive(n) {
			open = map[int]bool{}
		}
		open[page] = true
	}
	return map[string]string{PropOpen: formatOpen(open, len(Pages(n)))}
}

// Revealed is the properties that make a page shown (selecting one of its
// members does this); nil when it already is.
func Revealed(n design.Node, page int) map[string]string {
	if Shown(n, page) {
		return nil
	}
	if n.Component == design.TabsComponent {
		return map[string]string{PropActive: strconv.Itoa(page)}
	}
	open := OpenSet(n)
	if Exclusive(n) {
		open = map[int]bool{}
	}
	open[page] = true
	return map[string]string{PropOpen: formatOpen(open, len(Pages(n)))}
}

func formatOpen(open map[int]bool, count int) string {
	var parts []string
	for i := 0; i < count; i++ {
		if open[i] {
			parts = append(parts, strconv.Itoa(i))
		}
	}
	return strings.Join(parts, ",")
}

// Visible reports whether a component is drawn and can be hit as far as its
// container is concerned: a free component always is, a member only while its
// container is shown and its page is.
func Visible(doc design.Document, n design.Node) bool {
	if n.In == "" {
		return true
	}
	c, ok := doc.Get(n.In)
	if !ok || !c.IsContainer() {
		return true
	}
	return !c.Hidden && Shown(c, n.Page)
}

// Order is the indexes of the document's nodes in the order they are drawn,
// back to front: a member comes straight after its container, whatever its
// place in the list, so it is always above it. Hit testing walks it backwards.
func Order(doc design.Document) []int {
	members := map[design.NodeID][]int{}
	isMember := make([]bool, len(doc.Nodes))
	for i, n := range doc.Nodes {
		if n.In == "" {
			continue
		}
		if c, ok := doc.Get(n.In); ok && c.IsContainer() {
			members[n.In] = append(members[n.In], i)
			isMember[i] = true
		}
	}
	out := make([]int, 0, len(doc.Nodes))
	for i, n := range doc.Nodes {
		if isMember[i] {
			continue
		}
		out = append(out, i)
		out = append(out, members[n.ID]...)
	}
	return out
}

// PageAt is the page whose body holds the cell, among the pages that are
// shown, and whether there is one.
func PageAt(n design.Node, x, y int) (int, bool) {
	for i, b := range Layout(n).Bodies {
		if Shown(n, i) && b.H > 0 && b.Contains(x, y) {
			return i, true
		}
	}
	return 0, false
}

// Moves is how far the top-left of each page's body moved between two states
// of a container (it was moved, resized, or a section opened or closed): the
// members of a page move with its body.
func Moves(before, after design.Node) [][2]int {
	a, b := Layout(before).Bodies, Layout(after).Bodies
	out := make([][2]int, len(b))
	for i := range b {
		if i < len(a) {
			out[i] = [2]int{b[i].X - a[i].X, b[i].Y - a[i].Y}
		}
	}
	return out
}
