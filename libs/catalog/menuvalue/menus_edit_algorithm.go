package menuvalue

import "strings"

// AddMenu appends a menu with no entries. An empty name is "Menu".
func (m Menus) AddMenu(name string) Menus {
	out := m.clone()
	if name = strings.TrimSpace(name); name == "" {
		name = "Menu"
	}
	return append(out, Menu{Name: name, Entries: []Entry{}})
}

// RemoveMenu removes menu i with its entries.
func (m Menus) RemoveMenu(i int) Menus {
	if !m.hasMenu(i) {
		return m
	}
	out := m.clone()
	return append(out[:i], out[i+1:]...)
}

// RenameMenu gives menu i a name, which cannot be empty.
func (m Menus) RenameMenu(i int, name string) (Menus, error) {
	if name = strings.TrimSpace(name); name == "" {
		return m, ErrNoName
	}
	if !m.hasMenu(i) {
		return m, nil
	}
	out := m.clone()
	out[i].Name = name
	return out, nil
}

// MoveMenu moves menu from to index to, as the other menus close up around it.
func (m Menus) MoveMenu(from, to int) Menus {
	if !m.hasMenu(from) {
		return m
	}
	to = min(max(to, 0), len(m)-1)
	if from == to {
		return m
	}
	out := m.clone()
	mn := out[from]
	out = append(out[:from], out[from+1:]...)
	out = append(out[:to], append(Menus{mn}, out[to:]...)...)
	return out
}

// AddEntry appends an entry to menu i.
func (m Menus) AddEntry(menu int, e Entry) Menus {
	if !m.hasMenu(menu) {
		return m
	}
	out := m.clone()
	out[menu].Entries = append(out[menu].Entries, e)
	return out
}

// AddSeparator appends a separator to menu i.
func (m Menus) AddSeparator(menu int) Menus {
	return m.AddEntry(menu, Entry{Separator: true})
}

// RemoveEntry removes entry i of a menu.
func (m Menus) RemoveEntry(menu, i int) Menus {
	if !m.hasEntry(menu, i) {
		return m
	}
	out := m.clone()
	e := out[menu].Entries
	out[menu].Entries = append(e[:i], e[i+1:]...)
	return out
}

// SetEntry changes the name and shortcut of entry i. The name cannot be empty;
// a separator has neither and is left as it is.
func (m Menus) SetEntry(menu, i int, name, shortcut string) (Menus, error) {
	if !m.hasEntry(menu, i) || m[menu].Entries[i].Separator {
		return m, nil
	}
	if name = strings.TrimSpace(name); name == "" {
		return m, ErrNoName
	}
	out := m.clone()
	out[menu].Entries[i] = Entry{Name: name, Shortcut: strings.TrimSpace(shortcut)}
	return out, nil
}

// MoveEntry takes entry from of menu fromMenu out and puts it at index to of
// menu toMenu, counted after it was taken out, so moving to the end of its own
// menu is to == len-1. It may be another menu.
func (m Menus) MoveEntry(fromMenu, from, toMenu, to int) Menus {
	if !m.hasEntry(fromMenu, from) || !m.hasMenu(toMenu) {
		return m
	}
	out := m.clone()
	e := out[fromMenu].Entries[from]
	src := out[fromMenu].Entries
	out[fromMenu].Entries = append(src[:from], src[from+1:]...)
	dst := out[toMenu].Entries
	to = min(max(to, 0), len(dst))
	out[toMenu].Entries = append(dst[:to], append([]Entry{e}, dst[to:]...)...)
	return out
}

// NudgeEntry moves an entry one place up (dir -1) or down (+1). At the edge of
// its menu it crosses into the neighbouring menu, to the end of the one above
// or the start of the one below. It returns where the entry is now.
func (m Menus) NudgeEntry(menu, i, dir int) (out Menus, toMenu, to int) {
	if !m.hasEntry(menu, i) || dir == 0 {
		return m, menu, i
	}
	last := len(m[menu].Entries) - 1
	switch {
	case dir < 0 && i > 0:
		return m.MoveEntry(menu, i, menu, i-1), menu, i - 1
	case dir > 0 && i < last:
		return m.MoveEntry(menu, i, menu, i+1), menu, i + 1
	case dir < 0 && menu > 0:
		to = len(m[menu-1].Entries)
		return m.MoveEntry(menu, i, menu-1, to), menu - 1, to
	case dir > 0 && menu < len(m)-1:
		return m.MoveEntry(menu, i, menu+1, 0), menu + 1, 0
	}
	return m, menu, i
}
