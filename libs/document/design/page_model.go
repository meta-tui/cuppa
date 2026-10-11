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
	out := make([]Document, len(d.Pages))
	for i, p := range d.Pages {
		s := d
		s.Pages, s.Page = nil, 0
		s.Name = p.Name
		if i == active {
			s.Nodes = d.Nodes
		} else {
			s.Nodes = p.Nodes
		}
		out[i] = s.Clone()
	}
	return out
}

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
