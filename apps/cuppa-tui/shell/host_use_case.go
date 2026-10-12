package shell

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/runpane"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
	"github.com/meta-ui-labs/cuppa/libs/runhost/local"
)

// Hosting reports whether the stage shows the application the design draws,
// running, instead of the design.
func (m *Model) Hosting() bool { return m.hosting }

// toggleHost shows the running application, or goes back to editing. The
// application keeps running while the design is edited, so showing it again
// is instant; it is built again when the design is saved.
func (m *Model) toggleHost() {
	if m.hosting {
		m.hideHost()
		return
	}
	m.stopPreview()
	if m.ed.Dirty() {
		m.flow.Save()
		if m.ed.Dirty() || m.flow.Modal() != nil {
			// Saving was refused or asked something: nothing to run yet.
			return
		}
	}
	plan, prepare, ok := m.hostPlan()
	if !ok {
		return
	}
	w, h := m.ed.Document().Width, m.ed.Document().Height
	if m.host == nil {
		host, cmd := runpane.Start(plan, prepare, w, h)
		m.host = host
		m.pending = append(m.pending, cmd)
	} else if m.hostStale {
		m.pending = append(m.pending, m.host.Restart(plan, prepare, w, h))
	}
	m.hostStale = false
	m.hosting = true
	m.stg.SetRenderer(func() *grid.Grid { return m.host.Grid() })
	m.flow.SetStatus(m.host.Status())
}

// hostPlan is what to build and run for the design, or false after telling the
// person what is missing.
func (m *Model) hostPlan() (command.Plan, func() error, bool) {
	path := m.flow.Path()
	s := m.ed.RunSettings()
	if path != "" {
		merged, err := local.Load(path, s)
		if err != nil {
			m.flow.Notice("Cannot read the local run settings", err.Error())
			return command.Plan{}, nil, false
		}
		s = merged
	}
	if !s.IsSet() {
		m.flow.Notice("Nothing to run",
			"This design does not say how to run the application it draws.\n\n"+
				"Set it from a terminal:\n  cuppa run set <design.cuppa> --run \"<program> <args>\" [--build \"...\"]")
		return command.Plan{}, nil, false
	}
	dir := "."
	var prepare func() error
	if path != "" {
		dir = filepath.Dir(path)
		prepare = func() error {
			_, err := runpane.Regenerate(path)
			return err
		}
	}
	plan, err := command.PlanOf(s, dir)
	if err != nil {
		m.flow.Notice("Cannot run the application", err.Error())
		return command.Plan{}, nil, false
	}
	return plan, prepare, true
}

// hideHost goes back to editing; the application keeps running.
func (m *Model) hideHost() {
	if !m.hosting {
		return
	}
	m.hosting = false
	m.stg.SetRenderer(nil)
	m.flow.SetStatus("")
}

// stopHost ends the application and goes back to editing.
func (m *Model) stopHost() {
	m.hideHost()
	if m.host != nil {
		m.host.Stop()
		m.host = nil
	}
	m.hostStale = false
}

// rebuildHost builds the application again from the design just saved.
func (m *Model) rebuildHost() {
	if m.host == nil {
		return
	}
	plan, prepare, ok := m.hostPlan()
	if !ok {
		return
	}
	w, h := m.ed.Document().Width, m.ed.Document().Height
	m.pending = append(m.pending, m.host.Restart(plan, prepare, w, h))
	if m.hosting {
		m.flow.SetStatus(m.host.Status())
	}
}

// afterSave runs when the design was written: the running application is
// built again when the design says to do so on save, or when it is on show.
func (m *Model) afterSave() {
	if m.host == nil || m.ed.Dirty() {
		return
	}
	if m.hosting || m.ed.RunSettings().OnSave {
		m.rebuildHost()
		return
	}
	m.hostStale = true
}

// hostKey gives a key to the application; Ctrl+R goes back to editing and
// Ctrl+S saves and builds again. It reports whether the key was used.
func (m *Model) hostKey(msg tea.KeyPressMsg) bool {
	k := msg.Key()
	if k.Mod&tea.ModSuper != 0 {
		k.Mod = k.Mod&^tea.ModSuper | tea.ModCtrl
	}
	switch k.Keystroke() {
	case "ctrl+r":
		m.hideHost()
	case "ctrl+s":
		m.flow.Save()
		m.afterSave()
	case "ctrl+q":
		m.flow.Quit()
	case "ctrl+p", "ctrl+n", "ctrl+o", "ctrl+shift+s":
		return false
	default:
		m.host.Key(msg)
	}
	return true
}

// hostMouse hands pointer events over the stage to the application.
func (m *Model) hostMouse(e pointer.Event) {
	m.mouseX, m.mouseY = e.X, e.Y
	switch m.layout.paneAt(e.X, e.Y) {
	case inInspector:
		// The right bar is idle while the application runs, except for the
		// buttons that go back to editing.
		m.ins.HandleMode(e.Translate(m.layout.inspector.X, m.layout.inspector.Y))
	case inStage:
		if cx, cy, over := m.stageCell(e.X, e.Y); over {
			e.X, e.Y = cx, cy
			m.host.Mouse(e)
		}
	}
}

// hostMessage handles what the build and the program report.
func (m *Model) hostMessage(msg tea.Msg) {
	if m.host == nil {
		return
	}
	m.pending = append(m.pending, m.host.Update(msg))
	if m.hosting {
		m.flow.SetStatus(m.host.Status())
	}
}
