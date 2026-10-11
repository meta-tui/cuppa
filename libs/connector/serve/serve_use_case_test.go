package serve

import (
	"net"
	"testing"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/connector/message"
)

func listen(t *testing.T) *Server {
	t.Helper()
	s, err := Listen("127.0.0.1:0", Options{App: "demo", Pages: []string{"Run", "Help"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func connect(t *testing.T, s *Server) (*message.Reader, *message.Writer) {
	t.Helper()
	conn, err := net.Dial("tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	return message.NewReader(conn), message.NewWriter(conn)
}

func TestADesignerIsToldWhoIsListeningAndWhatEveryPageHolds(t *testing.T) {
	s := listen(t)
	if err := s.Publish("Run", map[string]any{"Title": "Run", "Width": 80}, true); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish("Help", map[string]any{"Title": "Help"}, false); err != nil {
		t.Fatal(err)
	}
	r, _ := connect(t, s)
	hello, err := r.Next()
	if err != nil || hello.Type != message.Hello || hello.App != "demo" || hello.Protocol != message.Protocol || len(hello.Pages) != 2 {
		t.Fatalf("hello: %+v, %v", hello, err)
	}
	for _, want := range []struct {
		page    string
		current bool
	}{{"Run", true}, {"Help", false}} {
		m, err := r.Next()
		if err != nil || m.Type != message.State || m.Page != want.page || m.Current != want.current {
			t.Fatalf("state of %s: %+v, %v", want.page, m, err)
		}
	}
}

func TestOnlyOnePageIsCurrent(t *testing.T) {
	s := listen(t)
	_ = s.Publish("Run", map[string]any{}, true)
	_ = s.Publish("Help", map[string]any{}, true)
	r, _ := connect(t, s)
	_, _ = r.Next()
	run, _ := r.Next()
	help, _ := r.Next()
	if run.Current || !help.Current {
		t.Fatalf("current: Run=%v Help=%v", run.Current, help.Current)
	}
}

func TestAChangeIsPublishedToTheDesigner(t *testing.T) {
	s := listen(t)
	r, _ := connect(t, s)
	if _, err := r.Next(); err != nil { // hello
		t.Fatal(err)
	}
	waitFor(t, func() bool { s.mu.Lock(); defer s.mu.Unlock(); return len(s.clients) == 1 })
	if err := s.Publish("Run", map[string]any{"Rows": []string{"a", "b"}}, true); err != nil {
		t.Fatal(err)
	}
	m, err := r.Next()
	if err != nil || m.Page != "Run" || string(m.Inputs["Rows"]) != `["a","b"]` {
		t.Fatalf("got %+v, %v", m, err)
	}
}

func TestAnEventFromTheDesignerReachesTheApplication(t *testing.T) {
	s := listen(t)
	r, w := connect(t, s)
	if _, err := r.Next(); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(message.Message{Type: message.Event, Page: "Run", Name: "Hint", X: 4}); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-s.Events():
		if e.Page != "Run" || e.Name != "Hint" || e.X != 4 {
			t.Fatalf("event %+v", e)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the event did not arrive")
	}
}

func TestOnlyThisComputerCanBeListenedOn(t *testing.T) {
	if _, err := Listen("0.0.0.0:0", Options{}); err == nil {
		t.Fatal("listening on every interface was allowed")
	}
}

func TestCloseDisconnectsTheDesigner(t *testing.T) {
	s := listen(t)
	r, _ := connect(t, s)
	if _, err := r.Next(); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	if _, err := r.Next(); err == nil {
		t.Fatal("the designer was not disconnected")
	}
	select {
	case <-s.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("not done")
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for i := 0; i < 500; i++ {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not reached")
}
