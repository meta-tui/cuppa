package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestScreenContractCommands(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.label", 0, 0)
	n, _ := e.Document().Get(id)

	if err := e.SetBinding(n.ID, "text", "Title"); err != nil {
		t.Fatal(err)
	}
	if got, _ := e.Document().Get(n.ID); got.Bind["text"] != "Title" {
		t.Fatalf("bind = %v", got.Bind)
	}
	if err := e.SetBinding(n.ID, "nope", "X"); err == nil {
		t.Error("unknown property accepted")
	}
	if err := e.SetBinding(n.ID, "text", "9bad"); err == nil {
		t.Error("bad name accepted")
	}
	if err := e.SetShowIf(n.ID, "Show status"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetEvent(n.ID, "Save"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetKeys("s=Save:save, esc=Back"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetKeys("s"); err == nil {
		t.Error("key without event accepted")
	}
	doc := e.Document()
	got, _ := doc.Get(n.ID)
	if got.ShowIf != "Show status" || got.Event != "Save" || len(doc.Keys) != 2 || doc.Keys[0].Label != "save" {
		t.Fatalf("node %+v keys %+v", got, doc.Keys)
	}

	// Each command is one undo step; clearing removes the binding.
	e.Undo()
	e.Undo()
	e.Undo()
	e.Undo()
	if got, _ := e.Document().Get(n.ID); got.Bind != nil {
		t.Errorf("undo left %v", got.Bind)
	}
	if err := e.SetBinding(n.ID, "text", "Title"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetBinding(n.ID, "text", ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := e.Document().Get(n.ID); got.Bind != nil {
		t.Errorf("clear left %v", got.Bind)
	}
	e.SetLocked(n.ID, true)
	if err := e.SetEvent(n.ID, "Save"); err == nil {
		t.Error("locked node accepted an event")
	}
	_ = design.Node{}
}
