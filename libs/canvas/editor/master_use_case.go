package editor

import (
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// The master page (ADR 0015) holds the components every page shows below its
// own. It is a page of the design that is not a screen: its components are
// added to each screen by Document.Screens, minus the ones a page hides.

// MasterPage is the index of the master page, or -1.
func (e *Editor) MasterPage() int { return e.doc.MasterIndex() }

// OnMaster reports whether the master is the page being edited.
func (e *Editor) OnMaster() bool {
	m := e.doc.MasterIndex()
	return m >= 0 && m == e.doc.ActivePage()
}

// AddMasterPage adds the master page and shows it, as one undo step. A design
// has one.
func (e *Editor) AddMasterPage() error {
	if e.doc.MasterIndex() >= 0 {
		return fmt.Errorf("editor: the design has a master page already")
	}
	e.apply(func() bool {
		e.pages()
		e.stashPage()
		e.doc.Pages = append(e.doc.Pages, design.Page{Name: e.uniquePageName("Master"), Master: true})
		e.showPage(len(e.doc.Pages) - 1)
		return true
	})
	return nil
}

// MasterComponents are the master's components, in z-order, as they were
// designed.
func (e *Editor) MasterComponents() []design.Node {
	return append([]design.Node(nil), e.doc.MasterNodes()...)
}

// MasterHidden reports whether page i does not show the master component.
func (e *Editor) MasterHidden(page int, id design.NodeID) bool {
	if page < 0 || page >= len(e.doc.Pages) {
		return false
	}
	for _, h := range e.doc.Pages[page].Hide {
		if h == id {
			return true
		}
	}
	return false
}

// SetMasterHidden makes page i hide a master component, or show it again, as
// one undo step.
func (e *Editor) SetMasterHidden(page int, id design.NodeID, hidden bool) error {
	if page < 0 || page >= len(e.doc.Pages) || e.doc.Pages[page].Master {
		return fmt.Errorf("editor: no such page")
	}
	known := false
	for _, n := range e.doc.MasterNodes() {
		if n.ID == id {
			known = true
		}
	}
	if !known {
		return fmt.Errorf("editor: the master has no such component")
	}
	if e.MasterHidden(page, id) == hidden {
		return nil
	}
	e.apply(func() bool {
		var hide []design.NodeID
		for _, h := range e.doc.Pages[page].Hide {
			if h != id {
				hide = append(hide, h)
			}
		}
		if hidden {
			hide = append(hide, id)
		}
		e.doc.Pages[page].Hide = hide
		return true
	})
	return nil
}

// MoveToMaster moves the selected components of the page being edited to the
// master, as one undo step, and says how many moved. A locked component and a
// container with its members stay where they are.
func (e *Editor) MoveToMaster() (int, error) {
	m := e.doc.MasterIndex()
	if m < 0 {
		return 0, fmt.Errorf("editor: the design has no master page; add one first")
	}
	if e.OnMaster() {
		return 0, fmt.Errorf("editor: these components are on the master already")
	}
	ids := e.unlockedSelection()
	moved := 0
	e.apply(func() bool {
		taken := map[string]bool{}
		for _, n := range e.doc.Pages[m].Nodes {
			taken[n.Name] = true
		}
		for _, id := range ids {
			n, ok := e.doc.Get(id)
			if !ok || n.IsContainer() || n.In != "" {
				continue
			}
			e.doc.Remove(id)
			n.ID = e.doc.NewID()
			for base, i := n.Name, 2; taken[n.Name]; i++ {
				n.Name = fmt.Sprintf("%s %d", base, i)
			}
			taken[n.Name] = true
			e.doc.Pages[m].Nodes = append(e.doc.Pages[m].Nodes, n)
			moved++
		}
		return moved > 0
	})
	e.sel = e.keepSelected(func(id design.NodeID) bool { return e.doc.Index(id) >= 0 })
	if moved == 0 {
		return 0, fmt.Errorf("editor: select components to move to the master")
	}
	return moved, nil
}

// RenderDocument is the design as it is drawn while editing a page: the
// master's components below the page's own, laid out for the canvas. They are
// not part of Document, so they cannot be selected or changed from the page.
func (e *Editor) RenderDocument() design.Document {
	doc := e.doc.Clone()
	if e.doc.MasterIndex() < 0 || e.OnMaster() {
		return doc
	}
	shown := e.doc.ShownMaster(e.doc.ActivePage())
	if len(shown) == 0 {
		return doc
	}
	doc.Nodes = append(shown, doc.Nodes...)
	laid := &Editor{cat: e.cat, doc: doc}
	laid.Resolve()
	return laid.doc
}
