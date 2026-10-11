// Package liveconn is the designer's end of the live connection (ADR 0016) as
// Bubble Tea commands: it dials a running application and turns what the
// application publishes into messages the shell applies to the editor.
package liveconn

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/libs/connector/dial"
)

// Conn is a connection to a running application.
type Conn struct {
	client *dial.Client
	addr   string
}

// ConnectedMsg is the result of Dial.
type ConnectedMsg struct {
	Conn *Conn
	Err  error
}

// StateMsg is the values an application published for one page.
type StateMsg struct {
	Conn    *Conn
	Page    string
	Current bool
	Inputs  map[string]json.RawMessage
}

// EndedMsg says the application closed the connection.
type EndedMsg struct{ Conn *Conn }

// Dial connects to addr, which must be on this computer, saying which design is
// open and what its pages are.
func Dial(addr, design string, pages []string) tea.Cmd {
	return func() tea.Msg {
		c, err := dial.Dial(addr, design, pages)
		if err != nil {
			return ConnectedMsg{Err: err}
		}
		return ConnectedMsg{Conn: &Conn{client: c, addr: addr}}
	}
}

// App is the name the application gave itself.
func (c *Conn) App() string {
	if name := c.client.Hello().App; name != "" {
		return name
	}
	return c.addr
}

// Pages are the pages the application says it can publish.
func (c *Conn) Pages() []string { return c.client.Hello().Pages }

// Wait is the command that delivers the next state, or EndedMsg when the
// connection is over. The shell asks for it again after each state.
func (c *Conn) Wait() tea.Cmd {
	return func() tea.Msg {
		m, ok := <-c.client.States()
		if !ok {
			return EndedMsg{Conn: c}
		}
		return StateMsg{Conn: c, Page: m.Page, Current: m.Current, Inputs: m.Inputs}
	}
}

// Send tells the application an event was raised on a page.
func (c *Conn) Send(page, name string, x, y int) error {
	return c.client.Send(page, name, x, y)
}

// Close ends the connection.
func (c *Conn) Close() { _ = c.client.Close() }
