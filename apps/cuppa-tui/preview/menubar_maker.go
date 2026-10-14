package preview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/packages/components/menubar"
)

func init() { makers["cuppa.menubar"] = makeMenuBar }

// makeMenuBar runs the menu bar component itself: pressing a menu opens its
// box, and the arrows and Enter work once it is open, as in a program that uses
// it.
func makeMenuBar(p props, w, h int) *live {
	m := menubar.New(
		menubar.WithMenusJSON(p.str("menus")),
		menubar.WithTitle(p.str("title")),
		menubar.WithOpen(p.integer("open", -1)),
		menubar.WithColours(p.str("color"), p.str("mutedColor"), p.str("ruleColor")),
		menubar.WithSize(w, h),
	)
	return &live{
		keys:   true,
		update: func(msg tea.Msg) tea.Cmd { var c tea.Cmd; m, c = m.Update(msg); return c },
		view:   func() string { return m.View() },
	}
}
