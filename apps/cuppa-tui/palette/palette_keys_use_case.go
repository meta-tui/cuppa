package palette

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// NavKey takes a key while the palette has the keyboard: up, down, home and
// end move the cursor, left and right fold and unfold a category, enter folds a
// category or returns the id of the component under the cursor, and "/" starts
// the search. The id is "" for every other key.
func (m *Model) NavKey(name string) (place string) {
	if len(m.rows) == 0 {
		if name == "/" {
			m.searching = true
		}
		return ""
	}
	m.cursor = min(max(m.cursor, 0), len(m.rows)-1)
	switch name {
	case "up":
		m.cursor--
	case "down":
		m.cursor++
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.rows) - 1
	case "left":
		m.fold(true)
	case "right":
		m.fold(false)
	case "enter":
		r := m.rows[m.cursor]
		if !r.header {
			return r.def.ID
		}
		m.fold(!m.collapsed[r.category])
	case "/":
		m.searching = true
	}
	m.cursor = min(max(m.cursor, 0), len(m.rows)-1)
	m.reveal()
	return ""
}

// fold folds or unfolds the category the cursor is in, and puts the cursor on its
// header so it stays where it was.
func (m *Model) fold(collapse bool) {
	if m.query != "" {
		return
	}
	r := m.rows[m.cursor]
	category := r.category
	if !r.header {
		category = r.def.Category
		if category == "" {
			category = definition.CategoryOther
		}
	}
	if m.collapsed[category] == collapse {
		return
	}
	m.collapsed[category] = collapse
	m.rebuild()
	for i, row := range m.rows {
		if row.header && row.category == category {
			m.cursor = i
		}
	}
	m.clampScroll()
}

// reveal scrolls so the cursor row is in view.
func (m *Model) reveal() {
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if h := m.listHeight(); h > 0 && m.cursor >= m.scroll+h {
		m.scroll = m.cursor - h + 1
	}
	m.clampScroll()
}
