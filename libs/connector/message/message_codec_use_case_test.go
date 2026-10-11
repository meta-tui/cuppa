package message

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestAMessageSurvivesTheWire(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	values, err := Values(map[string]any{"Title": "Run", "Width": 80, "Rows": []string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	in := Message{Type: State, Page: "Run", Current: true, Inputs: values}
	if err := w.Write(in); err != nil {
		t.Fatal(err)
	}
	if err := w.Write(Message{Type: Event, Page: "Run", Name: "ColoursSave", X: 3}); err != nil {
		t.Fatal(err)
	}
	if strings.Count(buf.String(), "\n") != 2 {
		t.Fatalf("not one line per message: %q", buf.String())
	}
	r := NewReader(&buf)
	got, err := r.Next()
	if err != nil || got.Type != State || got.Page != "Run" || !got.Current || string(got.Inputs["Title"]) != `"Run"` || string(got.Inputs["Width"]) != "80" {
		t.Fatalf("state: %+v, %v", got, err)
	}
	ev, err := r.Next()
	if err != nil || ev.Type != Event || ev.Name != "ColoursSave" || ev.X != 3 || ev.Y != 0 {
		t.Fatalf("event: %+v, %v", ev, err)
	}
	if _, err := r.Next(); !errors.Is(err, io.EOF) {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestALineThatIsNotAMessageIsSkipped(t *testing.T) {
	r := NewReader(strings.NewReader("not json\n\n{\"nothing\":1}\n{\"type\":\"hello\",\"protocol\":1}\n"))
	m, err := r.Next()
	if err != nil || m.Type != Hello || m.Protocol != Protocol {
		t.Fatalf("got %+v, %v", m, err)
	}
}

func TestALastLineWithoutANewlineIsRead(t *testing.T) {
	m, err := NewReader(strings.NewReader(`{"type":"state","page":"A"}`)).Next()
	if err != nil || m.Page != "A" {
		t.Fatalf("got %+v, %v", m, err)
	}
}
