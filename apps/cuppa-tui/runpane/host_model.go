package runpane

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/preview"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
	"github.com/meta-ui-labs/cuppa/libs/runhost/build"
	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
	"github.com/meta-ui-labs/cuppa/libs/runhost/session"
)

// coalesce is how long a burst of output waits before the pane is drawn
// again, so a program that prints fast does not draw a frame per write.
const coalesce = 16 * time.Millisecond

// Model is the program being hosted.
type Model struct {
	plan command.Plan
	// prepare runs before the build: the export of the saved design. It may be nil.
	prepare func() error
	w, h    int
	gen     int
	state   State
	output  string
	sess    *session.Session
	exit    error
	cancel  context.CancelFunc
}

// Start builds and runs the plan in a terminal of w by h cells (the plan's own
// fixed size wins). prepare runs before the build, in the background; its
// error is shown as a failed build. The returned command delivers the
// messages to give back to Update.
func Start(p command.Plan, prepare func() error, w, h int) (*Model, tea.Cmd) {
	m := &Model{plan: p, prepare: prepare}
	return m, m.launch(w, h)
}

// Restart stops the program and builds and runs it again.
func (m *Model) Restart(p command.Plan, prepare func() error, w, h int) tea.Cmd {
	m.halt()
	m.plan, m.prepare = p, prepare
	return m.launch(w, h)
}

func (m *Model) launch(w, h int) tea.Cmd {
	if m.plan.Width > 0 {
		w = m.plan.Width
	}
	if m.plan.Height > 0 {
		h = m.plan.Height
	}
	m.w, m.h = max(w, 1), max(h, 1)
	m.gen++
	m.output, m.exit = "", nil
	if m.plan.Build == nil && m.prepare == nil {
		return m.run()
	}
	m.state = Building
	gen, prepare, c := m.gen, m.prepare, m.plan.Build
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	return func() tea.Msg {
		if prepare != nil {
			if err := prepare(); err != nil {
				return BuiltMsg{Gen: gen, Err: err}
			}
		}
		if c == nil {
			return BuiltMsg{Gen: gen}
		}
		out, err := build.Run(ctx, *c)
		return BuiltMsg{Gen: gen, Out: out, Err: err}
	}
}

// run starts the program and waits for its first output.
func (m *Model) run() tea.Cmd {
	s, err := session.Start(m.plan.Run, m.w, m.h)
	if err != nil {
		m.state, m.output = BuildFailed, err.Error()
		return nil
	}
	m.sess, m.state = s, Running
	return m.wait()
}

func (m *Model) wait() tea.Cmd {
	s, gen := m.sess, m.gen
	return func() tea.Msg {
		select {
		case <-s.Changed():
			time.Sleep(coalesce)
		case <-s.Done():
		}
		return ChangedMsg{Gen: gen}
	}
}

// Update handles BuiltMsg and ChangedMsg; any other message is ignored.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case BuiltMsg:
		if msg.Gen != m.gen || m.state != Building {
			return nil
		}
		if msg.Err != nil {
			m.state = BuildFailed
			m.output = strings.TrimSpace(msg.Out + "\n" + msg.Err.Error())
			return nil
		}
		return m.run()
	case ChangedMsg:
		if msg.Gen != m.gen || m.sess == nil {
			return nil
		}
		select {
		case <-m.sess.Done():
			m.state, m.exit = Ended, m.sess.Err()
			return nil
		default:
		}
		return m.wait()
	}
	return nil
}

// State is what the pane is doing.
func (m *Model) State() State { return m.state }

// Size is the size of the program's terminal in cells.
func (m *Model) Size() (int, int) { return m.w, m.h }

// Status is one line saying what the pane is doing.
func (m *Model) Status() string {
	switch m.state {
	case Building:
		return "Building the app..."
	case BuildFailed:
		return "The build failed: Ctrl+S saves and builds again, Ctrl+R goes back to editing"
	case Ended:
		if m.exit != nil {
			return "The app ended (" + m.exit.Error() + "): Ctrl+S saves and runs it again, Ctrl+R goes back"
		}
		return "The app ended: Ctrl+S saves and runs it again, Ctrl+R goes back"
	}
	return "Running the app: your keys and mouse go to it; Ctrl+S saves and rebuilds, Ctrl+R goes back to editing"
}

// Grid is what the pane shows: the program's screen, or what the build said.
func (m *Model) Grid() *grid.Grid {
	switch m.state {
	case Running, Ended:
		if m.sess != nil {
			return preview.GridOf(m.sess.Render(), m.w, m.h)
		}
	case Building:
		return textGrid("Building the app...", m.w, m.h)
	}
	return textGrid(m.output, m.w, m.h)
}

// textGrid is lines of text on a grid, cut to its size.
func textGrid(text string, w, h int) *grid.Grid {
	g := grid.New(w, h)
	for y, line := range strings.Split(text, "\n") {
		if y >= h {
			break
		}
		for x, r := range []rune(strings.TrimRight(line, "\r")) {
			if x >= w {
				break
			}
			g.Set(x, y, grid.Cell{Ch: r})
		}
	}
	return g
}

// Resize follows the size of the canvas, unless the plan fixes the size.
func (m *Model) Resize(w, h int) {
	if m.plan.Width > 0 || m.plan.Height > 0 {
		return
	}
	m.w, m.h = max(w, 1), max(h, 1)
	if m.sess != nil {
		m.sess.Resize(m.w, m.h)
	}
}

// Key gives the program a key.
func (m *Model) Key(k tea.KeyPressMsg) {
	if m.sess != nil && m.state == Running {
		m.sess.SendKey(uv.KeyPressEvent(k))
	}
}

// Mouse gives the program a pointer event at a cell of its terminal.
func (m *Model) Mouse(e pointer.Event) {
	if m.sess == nil || m.state != Running {
		return
	}
	mouse := uv.Mouse{X: e.X, Y: e.Y}
	switch e.Phase {
	case pointer.Down:
		if e.Left {
			mouse.Button = uv.MouseLeft
			m.sess.SendMouse(uv.MouseClickEvent(mouse))
		}
	case pointer.Up:
		if e.Left {
			mouse.Button = uv.MouseLeft
			m.sess.SendMouse(uv.MouseReleaseEvent(mouse))
		}
	case pointer.Move:
		if e.Held {
			mouse.Button = uv.MouseLeft
		}
		m.sess.SendMouse(uv.MouseMotionEvent(mouse))
	case pointer.Wheel:
		switch {
		case e.WheelY < 0:
			mouse.Button = uv.MouseWheelUp
		case e.WheelY > 0:
			mouse.Button = uv.MouseWheelDown
		case e.WheelX < 0:
			mouse.Button = uv.MouseWheelLeft
		case e.WheelX > 0:
			mouse.Button = uv.MouseWheelRight
		default:
			return
		}
		m.sess.SendMouse(uv.MouseWheelEvent(mouse))
	}
}

// halt stops the build and the program, and makes late messages stale.
func (m *Model) halt() {
	m.gen++
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.sess != nil {
		m.sess.Stop()
		m.sess = nil
	}
}

// Stop ends the program.
func (m *Model) Stop() { m.halt() }

// String names the program, for messages.
func (m *Model) String() string {
	return fmt.Sprint(m.plan.Run.Argv)
}
