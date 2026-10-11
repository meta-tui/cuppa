package gosource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func contractDesign() design.Document {
	doc := slotsScreen()
	doc.Name = "Slots"
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Note", Rect: design.Rect{Y: 7, W: 10, H: 1}, ShowIf: "ShowNote", Event: "Open"})
	return doc
}

func TestTheContractListsInputsAndEventsByName(t *testing.T) {
	got := Contract(contractDesign(), standard.Default())
	if len(got) != 1 || got[0].Page != "Slots" {
		t.Fatalf("pages: %+v", got)
	}
	kinds := map[string]InputContract{}
	for _, in := range got[0].Inputs {
		kinds[in.Name] = in
	}
	if in := kinds["List"]; in.Kind != "rows" || strings.Join(in.Fields, ",") != "Mark,Name,Swatch,Hex,Style" {
		t.Errorf("List: %+v", in)
	}
	if kinds["ShowNote"].Kind != "bool" {
		t.Errorf("ShowNote: %+v", kinds["ShowNote"])
	}
	if _, ok := kinds["ListRows"]; ok {
		t.Error("the row source is not an input of the contract")
	}
	names := map[string]bool{}
	for _, e := range got[0].Events {
		names[e.Name] = true
	}
	if !names["Pick"] || !names["Open"] {
		t.Errorf("events: %+v", got[0].Events)
	}
}

const registryCheck = `package screens

import "testing"

func TestTheRegistryOfAScreen(t *testing.T) {
	in := SlotsInputs(DefaultSlotsProps())
	if _, ok := in["List"]; !ok {
		t.Fatalf("inputs %v", in)
	}
	if _, ok := in["ShowNote"].(bool); !ok {
		t.Fatalf("ShowNote %v", in["ShowNote"])
	}
	e, ok := SlotsEventNamed("Pick", 3, 2)
	if !ok || e != (SlotsPick{X: 3, Y: 2}) {
		t.Fatalf("event %v %v", e, ok)
	}
	if _, ok := SlotsEventNamed("Nope", 0, 0); ok {
		t.Fatal("an event that is not in the design")
	}
}
`

func TestAScreenWritesItsInputsByNameAndEventsFromNames(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	dir := filepath.Join(t.TempDir(), "screens")
	if err := WriteScreens(dir, GenerateScreens([]design.Document{contractDesign()}, standard.Default(), "screens")); err != nil {
		t.Fatal(err)
	}
	gomod := "module example.com/screens\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "registry_check_test.go"), []byte(registryCheck), 0o644); err != nil {
		t.Fatal(err)
	}
	goRun(t, dir, "mod", "tidy")
	goRun(t, dir, "vet", "./...")
	goRun(t, dir, "test", "./...")
}
