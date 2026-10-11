package scheme

import (
	"sort"
	"strings"
	"sync"

	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

var (
	once sync.Once
	all  []Scheme
)

// All returns every scheme, sorted by name without regard to case. The list is
// built once from bubbletint's default tints and must not be modified.
func All() []Scheme {
	once.Do(func() {
		for _, t := range tint.DefaultTints() {
			all = append(all, fromTint(t))
		}
		sort.SliceStable(all, func(i, j int) bool {
			return strings.ToLower(all[i].Name) < strings.ToLower(all[j].Name)
		})
	})
	return all
}

// Find returns the scheme with the given name (case-insensitive).
func Find(name string) (Scheme, bool) {
	for _, s := range All() {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}
	return Scheme{}, false
}

func fromTint(t *tint.Tint) Scheme {
	s := Scheme{Name: t.DisplayName, Dark: t.Dark}
	colours := [16]*tint.Color{
		t.Black, t.Red, t.Green, t.Yellow, t.Blue, t.Purple, t.Cyan, t.White,
		t.BrightBlack, t.BrightRed, t.BrightGreen, t.BrightYellow,
		t.BrightBlue, t.BrightPurple, t.BrightCyan, t.BrightWhite,
	}
	for i, c := range colours {
		if c == nil {
			s.ANSI[i] = space.ANSI(i) // a scheme with a gap falls back to the palette
			continue
		}
		s.ANSI[i] = rgb(c)
	}
	s.Foreground, s.Background = s.ANSI[7], s.ANSI[0]
	if t.Fg != nil {
		s.Foreground = rgb(t.Fg)
	}
	if t.Bg != nil {
		s.Background = rgb(t.Bg)
	}
	if t.Cursor != nil {
		s.Cursor, s.HasCursor = rgb(t.Cursor), true
	}
	if t.SelectionBg != nil {
		s.Selection, s.HasSelection = rgb(t.SelectionBg), true
	}
	return s
}

func rgb(c *tint.Color) space.RGB { return space.RGB{R: c.R, G: c.G, B: c.B} }
