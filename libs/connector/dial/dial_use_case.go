// Package dial is the designer's side of the live connection (ADR 0016): it
// connects to an application that listens, receives the values of its pages
// and sends back the events a click or key raises.
package dial

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/connector/address"
	"github.com/meta-ui-labs/cuppa/libs/connector/message"
)

// helloTimeout is how long an application has to say hello.
const helloTimeout = 3 * time.Second

// Client is a connection to an application.
type Client struct {
	conn  net.Conn
	out   *message.Writer
	hello message.Message
	in    chan message.Message
	done  chan struct{}
	once  sync.Once
}

// Dial connects to addr, which must be on this computer, and says which design
// is open and what its pages are. It returns when the application has said
// hello, or with an error when it did not or speaks another protocol.
func Dial(addr, design string, pages []string) (*Client, error) {
	if err := address.Loopback(addr); err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", addr, helloTimeout)
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, out: message.NewWriter(conn), in: make(chan message.Message, 64), done: make(chan struct{})}
	if err := c.out.Write(message.Message{Type: message.Hello, Protocol: message.Protocol, Design: design, Pages: pages}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	r := message.NewReader(conn)
	_ = conn.SetReadDeadline(time.Now().Add(helloTimeout))
	hello, err := r.Next()
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil || hello.Type != message.Hello {
		_ = conn.Close()
		if err == nil {
			err = errors.New("the application did not say hello")
		}
		return nil, err
	}
	if hello.Protocol != message.Protocol {
		_ = conn.Close()
		return nil, fmt.Errorf("the application speaks protocol %d, this designer %d", hello.Protocol, message.Protocol)
	}
	c.hello = hello
	go func() {
		defer close(c.in)
		for {
			m, err := r.Next()
			if err != nil {
				return
			}
			if m.Type == message.State {
				select {
				case c.in <- m:
				case <-c.done:
					return
				}
			}
		}
	}()
	return c, nil
}

// Hello is what the application said about itself: its name and its pages.
func (c *Client) Hello() message.Message { return c.hello }

// States are the states the application publishes, in order. It is closed
// when the connection ends.
func (c *Client) States() <-chan message.Message { return c.in }

// Send tells the application an event was raised on a page.
func (c *Client) Send(page, name string, x, y int) error {
	return c.out.Write(message.Message{Type: message.Event, Page: page, Name: name, X: x, Y: y})
}

// Close ends the connection.
func (c *Client) Close() error {
	c.once.Do(func() { close(c.done) })
	return c.conn.Close()
}
