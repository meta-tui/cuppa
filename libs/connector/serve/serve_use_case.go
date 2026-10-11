// Package serve is the application's side of the live connection (ADR 0016):
// it listens on this computer, tells each designer that connects which pages
// there are and what their inputs are now, publishes changes, and hands the
// events the designer sends back to the application.
package serve

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/connector/address"
	"github.com/meta-ui-labs/cuppa/libs/connector/message"
)

// writeTimeout is how long a designer may take to accept a message before it
// is dropped, so a stalled one cannot hold the application.
const writeTimeout = 2 * time.Second

// Options say who is listening.
type Options struct {
	// App is the application's name, shown in the designer.
	App string
	// Pages are the names of the pages the application can publish.
	Pages []string
}

// Server is a listening application.
type Server struct {
	ln   net.Listener
	opts Options

	mu      sync.Mutex
	clients map[*client]struct{}
	states  map[string]message.Message
	order   []string

	events chan message.Message
	done   chan struct{}
	once   sync.Once
}

type client struct {
	conn net.Conn
	out  *message.Writer
}

// Listen starts listening on addr, which must be on this computer.
func Listen(addr string, opts Options) (*Server, error) {
	if err := address.Loopback(addr); err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		ln: ln, opts: opts,
		clients: map[*client]struct{}{},
		states:  map[string]message.Message{},
		events:  make(chan message.Message, 64),
		done:    make(chan struct{}),
	}
	go s.accept()
	return s, nil
}

// Addr is the address the server listens on (useful when the port was 0).
func (s *Server) Addr() string { return s.ln.Addr().String() }

// Events are the events the designer sends, in order.
func (s *Server) Events() <-chan message.Message { return s.events }

// Done is closed when the server has been closed.
func (s *Server) Done() <-chan struct{} { return s.done }

// Publish sets the inputs of a page, by the input names of the design, and
// sends them to every designer connected. current says the application is
// showing that page. A designer that connects later is sent the latest state
// of every page.
func (s *Server) Publish(page string, inputs map[string]any, current bool) error {
	raw, err := message.Values(inputs)
	if err != nil {
		return err
	}
	m := message.Message{Type: message.State, Page: page, Current: current, Inputs: raw}
	s.mu.Lock()
	if _, known := s.states[page]; !known {
		s.order = append(s.order, page)
	}
	if current {
		for name, st := range s.states {
			if st.Current {
				st.Current = false
				s.states[name] = st
			}
		}
	}
	s.states[page] = m
	clients := make([]*client, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()
	for _, c := range clients {
		s.send(c, m)
	}
	return nil
}

// Close stops listening and disconnects every designer.
func (s *Server) Close() error {
	var err error
	s.once.Do(func() {
		close(s.done)
		err = s.ln.Close()
		s.mu.Lock()
		for c := range s.clients {
			_ = c.conn.Close()
		}
		s.mu.Unlock()
	})
	return err
}

func (s *Server) send(c *client, m message.Message) {
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := c.out.Write(m); err != nil {
		s.drop(c)
	}
}

func (s *Server) drop(c *client) {
	_ = c.conn.Close()
	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()
}

func (s *Server) accept() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go s.serveConn(conn)
	}
}

func (s *Server) serveConn(conn net.Conn) {
	c := &client{conn: conn, out: message.NewWriter(conn)}
	// The newcomer is told who is listening and what every page holds now,
	// before it can receive a later change.
	s.mu.Lock()
	first := []message.Message{{Type: message.Hello, Protocol: message.Protocol, App: s.opts.App, Pages: s.opts.Pages}}
	for _, page := range s.order {
		first = append(first, s.states[page])
	}
	var err error
	for _, m := range first {
		_ = conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err = c.out.Write(m); err != nil {
			break
		}
	}
	if err == nil {
		s.clients[c] = struct{}{}
	}
	s.mu.Unlock()
	if err != nil {
		_ = conn.Close()
		return
	}
	defer s.drop(c)

	r := message.NewReader(conn)
	for {
		m, err := r.Next()
		if err != nil {
			return
		}
		if m.Type != message.Event {
			continue
		}
		select {
		case s.events <- m:
		case <-s.done:
			return
		}
	}
}
