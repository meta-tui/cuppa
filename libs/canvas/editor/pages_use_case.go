package editor

import (
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// PageNames are the names of the design's pages in order.
func (e *Editor) PageNames() []string { return e.doc.PageNames() }

// ActivePage is the index of the page being edited.
func (e *Editor) ActivePage() int { return e.doc.ActivePage() }

// pages makes the design paged: a design with one page becomes a design with
// one named page, the one being edited.
func (e *Editor) pages() {
	if e.doc.IsPaged() {
		return
	}
	e.doc.PerPage = true
	e.doc.Pages = []design.Page{{Name: "Page 1"}}
	e.doc.Page = 0
}

// uniquePageName returns name, or "name 2", "name 3"... when it is taken.
func (e *Editor) uniquePageName(name string) string {
	taken := map[string]bool{}
	for _, n := range e.doc.PageNames() {
		taken[n] = true
	}
	if !taken[name] {
		return name
	}
	for i := 2; ; i++ {
		if candidate := fmt.Sprintf("%s %d", name, i); !taken[candidate] {
			return candidate
		}
	}
}

// AddPage adds an empty page after the others and shows it, as one undo step.
// An empty name is "Page N".
func (e *Editor) AddPage(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		name = fmt.Sprintf("Page %d", e.doc.PageCount()+1)
	} else if err := design.ValidPageName(name); err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	e.apply(func() bool {
		e.pages()
		e.stashPage()
		e.doc.Pages = append(e.doc.Pages, design.Page{Name: e.uniquePageName(name)})
		e.showPage(len(e.doc.Pages) - 1)
		return true
	})
	return nil
}

// stashPage puts the components of the page being edited back into its page.
func (e *Editor) stashPage() {
	if e.doc.IsPaged() {
		e.doc.Pages[e.doc.ActivePage()].Nodes = e.doc.Nodes
		e.doc.Pages[e.doc.ActivePage()].Samples = e.doc.Samples
		e.doc.Pages[e.doc.ActivePage()].Keys = e.doc.Keys
	}
}

// showPage makes page i the one being edited.
func (e *Editor) showPage(i int) {
	e.doc.Nodes = e.doc.Pages[i].Nodes
	if e.doc.Nodes == nil {
		e.doc.Nodes = []design.Node{}
	}
	e.doc.Pages[i].Nodes = nil
	e.doc.Samples = e.doc.Pages[i].Samples
	e.doc.Pages[i].Samples = nil
	e.doc.Keys = e.doc.Pages[i].Keys
	e.doc.Pages[i].Keys = nil
	e.doc.Page = i
	e.sel = nil
	// The canvas may have been resized while the page was away.
	e.Resolve()
}

// SwitchPage shows page i for editing. It is not an undo step: it is where
// the person is, not something they did to the design.
func (e *Editor) SwitchPage(i int) bool {
	if !e.doc.IsPaged() || i < 0 || i >= len(e.doc.Pages) || i == e.doc.ActivePage() {
		return false
	}
	e.stashPage()
	e.showPage(i)
	return true
}

// RenamePage gives page i another name, as one undo step; the components that
// go to it follow.
func (e *Editor) RenamePage(i int, name string) error {
	name = strings.TrimSpace(name)
	if !e.doc.IsPaged() || i < 0 || i >= len(e.doc.Pages) {
		return fmt.Errorf("editor: no such page")
	}
	if err := design.ValidPageName(name); err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	old := e.doc.Pages[i].Name
	if name == old {
		return nil
	}
	for j, p := range e.doc.Pages {
		if j != i && p.Name == name {
			return fmt.Errorf("editor: there is already a page called %q", name)
		}
	}
	e.apply(func() bool {
		e.doc.Pages[i].Name = name
		e.retarget(old, name)
		return true
	})
	return nil
}

// PageFolder is the folder an export keeps page i in ("" for the default).
func (e *Editor) PageFolder(i int) string {
	if !e.doc.IsPaged() || i < 0 || i >= len(e.doc.Pages) {
		return ""
	}
	return e.doc.Pages[i].Folder
}

// SetPageFolder sets the folder an export keeps page i in, as one undo step.
// An empty folder is pages/<page>. The design is made paged when it had one
// page only.
func (e *Editor) SetPageFolder(i int, folder string) error {
	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if err := design.ValidFolder(folder); err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	if i != 0 && (!e.doc.IsPaged() || i < 0 || i >= len(e.doc.Pages)) {
		return fmt.Errorf("editor: no such page")
	}
	if e.doc.IsPaged() && e.doc.Pages[i].Folder == folder {
		return nil
	}
	e.apply(func() bool {
		e.pages()
		e.doc.Pages[i].Folder = folder
		return true
	})
	return nil
}

// DeletePage removes page i with its components, as one undo step. A design
// keeps at least one page; the components that went to it no longer go
// anywhere.
func (e *Editor) DeletePage(i int) error {
	if !e.doc.IsPaged() || i < 0 || i >= len(e.doc.Pages) {
		return fmt.Errorf("editor: no such page")
	}
	if len(e.doc.Pages) < 2 {
		return fmt.Errorf("editor: a design keeps at least one page")
	}
	e.apply(func() bool {
		e.stashPage()
		gone := e.doc.Pages[i].Name
		pages := append(append([]design.Page(nil), e.doc.Pages[:i]...), e.doc.Pages[i+1:]...)
		active := e.doc.ActivePage()
		switch {
		case active == i:
			active = min(i, len(pages)-1)
		case active > i:
			active--
		}
		e.doc.Pages = pages
		e.retarget(gone, "")
		e.showPage(active)
		if len(e.doc.Pages) == 1 {
			// One page left is no split: it is the design again.
			e.doc.Pages, e.doc.Page = nil, 0
		}
		return true
	})
	return nil
}

// MovePage moves page i to position to, as one undo step.
func (e *Editor) MovePage(i, to int) bool {
	n := len(e.doc.Pages)
	if !e.doc.IsPaged() || i < 0 || i >= n || to < 0 || to >= n || i == to {
		return false
	}
	return e.apply(func() bool {
		e.stashPage()
		active := e.doc.Pages[e.doc.ActivePage()].Name
		p := e.doc.Pages[i]
		rest := append(append([]design.Page(nil), e.doc.Pages[:i]...), e.doc.Pages[i+1:]...)
		out := append(append(append([]design.Page(nil), rest[:to]...), p), rest[to:]...)
		e.doc.Pages = out
		for j, q := range out {
			if q.Name == active {
				e.showPage(j)
			}
		}
		return true
	})
}

// SetGoto sets the page a click on the component shows ("" for none), as one
// undo step.
func (e *Editor) SetGoto(id design.NodeID, page string) error {
	n, ok := e.doc.Get(id)
	if !ok {
		return fmt.Errorf("editor: no such component")
	}
	if n.Locked {
		return fmt.Errorf("editor: %s is locked", n.Name)
	}
	if page != "" && e.doc.PageIndex(page) < 0 {
		return fmt.Errorf("editor: there is no page called %q", page)
	}
	if n.Goto == page {
		return nil
	}
	e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Goto = page })
	})
	return nil
}

// retarget changes the components that go to page old so they go to to (or to
// nothing, for ""), on every page.
func (e *Editor) retarget(old, to string) {
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for i := range nodes {
			if nodes[i].Goto == old {
				nodes[i].Goto = to
			}
			walk(nodes[i].Children)
		}
	}
	walk(e.doc.Nodes)
	for i := range e.doc.Pages {
		walk(e.doc.Pages[i].Nodes)
	}
}

// ExportSettings is where the design says its Go source goes.
func (e *Editor) ExportSettings() design.ExportSettings { return e.doc.Export }

// SetExportSettings records where the design's Go source goes, as one undo
// step: the design is the one file that knows the structure of the program.
func (e *Editor) SetExportSettings(s design.ExportSettings) error {
	s = s.Clean()
	if s.Layout == "" {
		return fmt.Errorf("editor: an export needs a layout and a folder")
	}
	if s.Library && s.Module == "" {
		return fmt.Errorf("editor: a library export needs the import path of its folder")
	}
	if e.doc.Export == s {
		return nil
	}
	e.apply(func() bool {
		e.doc.Export = s
		return true
	})
	return nil
}
