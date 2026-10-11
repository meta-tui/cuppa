package shell

import (
	"encoding/json"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/liveconn"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
)

// defaultLiveAddress is what Connect… suggests: the address the documentation
// uses for an application that listens.
const defaultLiveAddress = "127.0.0.1:7331"

// askConnect asks for the address of a running application and connects.
func (m *Model) askConnect() {
	m.flow.Show(prompt.New("Connect to a running app", "Address on this computer", defaultLiveAddress), func(o modal.Outcome) {
		if !o.Canceled {
			m.Connect(strings.TrimSpace(o.Value))
		}
	})
}

// Connect dials the application listening on addr (this computer only). The
// result arrives as a message; nothing about it is saved in the design.
func (m *Model) Connect(addr string) {
	if addr == "" {
		return
	}
	m.disconnect("")
	m.pending = append(m.pending, liveconn.Dial(addr, m.ed.Document().Name, m.ed.PageNames()))
}

// disconnect ends the connection and goes back to the values of the design.
// why is shown when it is not empty.
func (m *Model) disconnect(why string) {
	if m.live != nil {
		m.live.Close()
		m.live = nil
	}
	m.ed.ClearLive()
	m.liveStates = nil
	if m.run != nil {
		m.pending = append(m.pending, m.run.ClearLive())
	}
	if why != "" {
		m.flow.Notice("Disconnected", why)
	}
}

// liveMessage applies what the connection delivered.
func (m *Model) liveMessage(msg any) {
	switch msg := msg.(type) {
	case liveconn.ConnectedMsg:
		if msg.Err != nil {
			m.flow.Notice("Cannot connect", msg.Err.Error()+"\n\nThe application must listen on this computer (ADR 0016); start it with its live option and use the address it prints.")
			return
		}
		m.live = msg.Conn
		m.pending = append(m.pending, msg.Conn.Wait())
	case liveconn.StateMsg:
		if msg.Conn != m.live {
			return
		}
		page := msg.Page
		names := m.ed.PageNames()
		known := false
		for _, n := range names {
			known = known || n == page
		}
		if !known && len(m.ed.Document().Pages) == 0 {
			// A design of one page shows the application's page, whatever it is called.
			page = names[0]
			known = true
		}
		if known {
			m.ed.SetLive(page, msg.Inputs)
			if m.liveStates == nil {
				m.liveStates = map[string]map[string]json.RawMessage{}
			}
			m.liveStates[page] = msg.Inputs
			if m.run != nil {
				m.pending = append(m.pending, m.run.SetLive(page, msg.Inputs))
			}
			if msg.Current {
				for i, n := range names {
					if n == page && i != m.ed.ActivePage() {
						m.ed.SwitchPage(i)
					}
				}
			}
		}
		m.pending = append(m.pending, msg.Conn.Wait())
	case liveconn.EndedMsg:
		if msg.Conn == m.live {
			m.disconnect("The application closed the connection; the designer shows the values of the design again.")
		}
	}
}

// sendLiveEvent tells the connected application that a click or a key raised
// an event on the page being run. It is sent from a command so a slow
// application cannot hold the designer.
func (m *Model) sendLiveEvent(name string, x, y int) {
	if m.live == nil || m.run == nil {
		return
	}
	conn, page := m.live, m.run.PageName()
	m.pending = append(m.pending, func() tea.Msg {
		_ = conn.Send(page, name, x, y)
		return nil
	})
}
