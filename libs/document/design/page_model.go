package design

import (
	"fmt"
	"strings"
)

// Page is one screen of a design that has several. Page is how a design is
// split into the screens of an application: each has its own components, and
// a component can take the person to another page (Node.Goto).
//
// Only the pages that are not shown keep their components here; the page that
// is shown has them in Document.Nodes, so every editing operation works on
// the shown page as it does on a design with one page.
type Page struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes,omitempty"`
	// Folder is where an export keeps the page, relative to the export's
	// root ("setup/preferences"); empty for pages/<page>. See ADR 0013.
	Folder string `json:"folder,omitempty"`
	// Samples are the layout input samples of the page when it is not the one
	// being edited; the shown page's are Document.Samples. Two pages may read
	// the same input with different values.
	Samples map[string]string `json:"samples,omitempty"`
	// Defaults are the app defaults of the page when it is not the one being
	// edited; the shown page's are Document.Defaults.
	Defaults map[string]string `json:"defaults,omitempty"`
	// Keys are the keys of the page when it is not the one being edited; the
	// shown page's are Document.Keys.
	Keys []KeyBinding `json:"keys,omitempty"`
	// Master marks the page whose components every other page shows below its
	// own (ADR 0015). There is at most one, and it is not a screen.
	Master bool `json:"master,omitempty"`
	// Hide are the master components this page does not show.
	Hide []NodeID `json:"hide,omitempty"`
	// Values give, for this page, the text of a property of a master component
	// that is bound to an input: input name to value. The master has one value
	// for it; this page previews and starts from its own.
	Values map[string]string `json:"values,omitempty"`
}

// PageCount is how many pages the design has: at least one.
func (d Document) PageCount() int { return max(len(d.Pages), 1) }

// IsPaged reports whether the design has been split into pages.
func (d Document) IsPaged() bool { return len(d.Pages) > 0 }

// ActivePage is the page being edited, kept inside the pages.
func (d Document) ActivePage() int {
	return min(max(d.Page, 0), d.PageCount()-1)
}

// PageNames are the names of the pages in order; a design with one page has
// the one name of the design.
func (d Document) PageNames() []string {
	if !d.IsPaged() {
		return []string{d.Name}
	}
	names := make([]string, len(d.Pages))
	for i, p := range d.Pages {
		names[i] = p.Name
	}
	return names
}

// PageIndex is the index of the page called name, or -1.
func (d Document) PageIndex(name string) int {
	for i, n := range d.PageNames() {
		if n == name {
			return i
		}
	}
	return -1
}

// Screens are the pages as designs of their own, in order: each has the
// components of its page and what the pages share (canvas, theme, keys,
// embedded components). A design with one page is its own only screen. A page
// is named after itself, so the exported screen is.
func (d Document) Screens() []Document {
	if !d.IsPaged() {
		return []Document{d}
	}
	active := d.ActivePage()
	var out []Document
	for i, p := range d.Pages {
		if p.Master {
			continue
		}
		s := d
		s.Pages, s.Page = nil, 0
		s.Name = p.Name
		s.Folder = p.Folder
		if i == active {
			s.Nodes = d.Nodes
		} else {
			s.Nodes = p.Nodes
			s.Samples = p.Samples
			s.Defaults = p.Defaults
			s.Keys = p.Keys
		}
		s.Nodes = append(d.ShownMaster(i), s.Nodes...)
		out = append(out, s.Clone())
	}
	if len(out) == 0 {
		// A design of nothing but its master is that master.
		s := d
		s.Pages, s.Page = nil, 0
		out = append(out, s.Clone())
	}
	return out
}

// MasterIndex is the index of the master page, or -1 when the design has none.
func (d Document) MasterIndex() int {
	for i, p := range d.Pages {
		if p.Master {
			return i
		}
	}
	return -1
}

// MasterNodes are the components of the master page, as they were designed.
func (d Document) MasterNodes() []Node {
	m := d.MasterIndex()
	switch {
	case m < 0:
		return nil
	case m == d.ActivePage():
		return d.Nodes
	}
	return d.Pages[m].Nodes
}

// MasterPrefix starts the id of a master component as a page shows it, so it
// never takes the id of one of the page's own.
const MasterPrefix = "m_"

// ShownMaster are the master's components page i shows: the ones it does not
// hide, with ids that carry MasterPrefix. The page that is the master shows
// none of them, as it is where they are made.
func (d Document) ShownMaster(i int) []Node {
	m := d.MasterIndex()
	if m < 0 || i == m || i < 0 || i >= len(d.Pages) {
		return nil
	}
	hidden := map[NodeID]bool{}
	for _, id := range d.Pages[i].Hide {
		hidden[id] = true
	}
	var out []Node
	for _, n := range d.MasterNodes() {
		if hidden[n.ID] {
			continue
		}
		n = prefixed(n)
		for prop, input := range n.Bind {
			if v, ok := d.Pages[i].Values[input]; ok {
				props := make(map[string]string, len(n.Props)+1)
				for k, pv := range n.Props {
					props[k] = pv
				}
				props[prop] = v
				n.Props = props
			}
		}
		out = append(out, n)
	}
	return out
}

func prefixed(n Node) Node {
	n = n.Clone()
	n.ID = MasterPrefix + n.ID
	if n.In != "" {
		n.In = MasterPrefix + n.In
	}
	for i := range n.Children {
		n.Children[i] = prefixed(n.Children[i])
	}
	return n
}

// ScreenCount is how many screens the design exports: its pages, without the
// master.
func (d Document) ScreenCount() int { return len(d.Screens()) }

// ValidPageName reports whether a page name can be used: it becomes the name
// of a Go screen, so it is letters, digits and spaces, and starts with a
// letter.
func ValidPageName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("a page needs a name")
	}
	if len([]rune(name)) > 40 {
		return fmt.Errorf("a page name is at most 40 characters")
	}
	for i, r := range name {
		letter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		digit := r >= '0' && r <= '9'
		if i == 0 && !letter || !letter && !digit && r != ' ' && r != '-' && r != '_' {
			return fmt.Errorf("a page name is letters, digits, spaces, - and _, and starts with a letter")
		}
	}
	return nil
}

// ValidFolder reports whether a folder can be used for a page: relative, in
// slash form, with lower case letters, digits, - and _ in each element.
func ValidFolder(folder string) error {
	if folder == "" {
		return nil
	}
	if len(folder) > 80 {
		return fmt.Errorf("a folder is at most 80 characters")
	}
	for _, el := range strings.Split(folder, "/") {
		if el == "" || el == "." || el == ".." || el == "internal" {
			return fmt.Errorf("a folder is names joined by /, such as setup/preferences")
		}
		for _, r := range el {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' {
				return fmt.Errorf("a folder uses lower case letters, digits, - and _")
			}
		}
	}
	return nil
}
