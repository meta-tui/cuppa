// Package menuvalue is the value of a "menu" property: a list of menus, each
// with entries that have a name and an optional shortcut, or are separators.
// It is stored in a design as JSON text, and the details bar edits it with the
// operations here, each of which returns a new value and leaves the old one
// alone.
package menuvalue

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Entry is one line of a menu: a named entry with an optional shortcut, or a
// separator.
type Entry struct {
	Name      string `json:"name,omitempty"`
	Shortcut  string `json:"shortcut,omitempty"`
	Separator bool   `json:"separator,omitempty"`
}

// Menu is one heading of the bar and its entries.
type Menu struct {
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

// Menus is the whole value.
type Menus []Menu

// ErrNoName is returned when a menu or an entry would have no name.
var ErrNoName = errors.New("a name is needed")

// Parse reads the stored text. Empty text is no menus.
func Parse(text string) (Menus, error) {
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	var menus Menus
	if err := json.Unmarshal([]byte(text), &menus); err != nil {
		return nil, fmt.Errorf("menus are not a list of menus: %w", err)
	}
	for i, m := range menus {
		if strings.TrimSpace(m.Name) == "" {
			return nil, fmt.Errorf("menu %d has no name", i+1)
		}
		for j, e := range m.Entries {
			if !e.Separator && strings.TrimSpace(e.Name) == "" {
				return nil, fmt.Errorf("entry %d of %q has no name", j+1, m.Name)
			}
		}
	}
	return menus, nil
}

// Encode writes the stored text: a JSON list, with no entries written as [].
func (m Menus) Encode() string {
	out := make(Menus, len(m))
	for i, mn := range m {
		out[i] = Menu{Name: mn.Name, Entries: append([]Entry{}, mn.Entries...)}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// Normalise parses text and writes it back in its canonical form.
func Normalise(text string) (string, error) {
	menus, err := Parse(text)
	if err != nil {
		return "", err
	}
	return menus.Encode(), nil
}

// clone copies the menus and their entries.
func (m Menus) clone() Menus {
	out := make(Menus, len(m))
	for i, mn := range m {
		out[i] = Menu{Name: mn.Name, Entries: append([]Entry(nil), mn.Entries...)}
	}
	return out
}

func (m Menus) hasMenu(i int) bool { return i >= 0 && i < len(m) }

func (m Menus) hasEntry(menu, i int) bool {
	return m.hasMenu(menu) && i >= 0 && i < len(m[menu].Entries)
}
