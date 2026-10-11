package dial

import (
	"net"
	"testing"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/connector/message"
	"github.com/meta-ui-labs/cuppa/libs/connector/serve"
)

// A fake application: the real server, driven the way a program would.
func TestTheDesignerSeesTheValuesAndTheApplicationSeesTheEvents(t *testing.T) {
	app, err := serve.Listen("127.0.0.1:0", serve.Options{App: "MVD", Pages: []string{"Run"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = app.Close() }()
	if err := app.Publish("Run", map[string]any{"Title": "Run", "Count": 3}, true); err != nil {
		t.Fatal(err)
	}

	c, err := Dial(app.Addr(), "mvd", []string{"Run", "Help"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	if h := c.Hello(); h.App != "MVD" || len(h.Pages) != 1 {
		t.Fatalf("hello %+v", h)
	}
	m := next(t, c)
	if m.Page != "Run" || !m.Current || string(m.Inputs["Count"]) != "3" {
		t.Fatalf("state %+v", m)
	}

	// The application changes a value; the designer sees it.
	if err := app.Publish("Run", map[string]any{"Title": "Run", "Count": 4}, true); err != nil {
		t.Fatal(err)
	}
	if m := next(t, c); string(m.Inputs["Count"]) != "4" {
		t.Fatalf("state %+v", m)
	}

	// A click in the designer reaches the application.
	if err := c.Send("Run", "Start", 2, 1); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-app.Events():
		if e.Name != "Start" || e.Page != "Run" || e.X != 2 || e.Y != 1 {
			t.Fatalf("event %+v", e)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no event")
	}
}

func TestTheStatesEndWhenTheApplicationGoes(t *testing.T) {
	app, err := serve.Listen("127.0.0.1:0", serve.Options{App: "x"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := Dial(app.Addr(), "d", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Close() }()
	_ = app.Close()
	select {
	case _, open := <-c.States():
		if open {
			t.Fatal("unexpected state")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the states did not end")
	}
}

func TestAnApplicationThatIsNotThereIsAnError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	if _, err := Dial(addr, "d", nil); err == nil {
		t.Fatal("dialled nothing")
	}
}

func TestAnotherProtocolIsRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = message.NewWriter(conn).Write(message.Message{Type: message.Hello, Protocol: 99})
		time.Sleep(time.Second)
	}()
	if _, err := Dial(ln.Addr().String(), "d", nil); err == nil {
		t.Fatal("protocol 99 was accepted")
	}
}

func TestOnlyThisComputerCanBeDialled(t *testing.T) {
	if _, err := Dial("example.com:80", "d", nil); err == nil {
		t.Fatal("a remote address was dialled")
	}
}

func next(t *testing.T, c *Client) message.Message {
	t.Helper()
	select {
	case m, ok := <-c.States():
		if !ok {
			t.Fatal("the states ended")
		}
		return m
	case <-time.After(5 * time.Second):
		t.Fatal("no state")
	}
	return message.Message{}
}
