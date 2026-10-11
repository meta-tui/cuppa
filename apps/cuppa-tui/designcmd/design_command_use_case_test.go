package designcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const sample = `{
  "document": {
    "name": "Hello", "width": 40, "height": 6,
    "keys": [{"key": "esc", "event": "Back", "label": "back"}],
    "nodes": [
      {"id": "n1", "component": "lipgloss.label", "name": "Title", "rect": {"X": 1, "Y": 1, "W": 20, "H": 1},
       "props": {"text": "Hi"}, "bind": {"text": "Title"}, "event": "Open"}
    ]
  }
}`

func run(t *testing.T, args ...string) (code int, out, errs string) {
	t.Helper()
	var o, e bytes.Buffer
	code = Run(args, &o, &e)
	return code, o.String(), e.String()
}

func TestADesignIsBuiltFromJSONAndReadBack(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "hello.json")
	if err := os.WriteFile(src, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "hello.cuppa")
	if code, o, e := run(t, "build", src, out); code != 0 || !strings.Contains(o, "Hello") {
		t.Fatalf("build: %d\n%s%s", code, o, e)
	}
	if code, o, e := run(t, "json", out); code != 0 || !strings.Contains(o, `"bind"`) || !strings.Contains(o, `"Title"`) {
		t.Fatalf("json: %d\n%s%s", code, o, e)
	}
	if code, o, e := run(t, "check", out); code != 0 || !strings.Contains(o, "in order") {
		t.Fatalf("check: %d\n%s%s", code, o, e)
	}
}

func TestAMisspeltFieldAndAnUnknownComponentAreRefused(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(src, []byte(strings.Replace(sample, `"bind"`, `"bnid"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := run(t, "build", src); code != 1 || !strings.Contains(e, "bnid") {
		t.Errorf("a misspelt field: %d %s", code, e)
	}
	if err := os.WriteFile(src, []byte(strings.Replace(sample, "lipgloss.label", "lipgloss.nothing", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := run(t, "build", src); code != 1 || !strings.Contains(e, "unknown component") {
		t.Errorf("an unknown component: %d %s", code, e)
	}
	if err := os.WriteFile(src, []byte(strings.Replace(sample, `"text": "Hi"`, `"colour": "red"`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := run(t, "build", src); code != 1 || !strings.Contains(e, `no property "colour"`) {
		t.Errorf("an unknown property: %d %s", code, e)
	}
	if code, _, _ := run(t, "nonsense", "x"); code != 2 {
		t.Errorf("usage exit %d", code)
	}
}

func TestContractPrintsInputsAndEventsAsJSON(t *testing.T) {
	doc := design.NewDocument("Run", 40, 6)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 3},
		Props: map[string]string{"title": "x"}, Bind: map[string]string{"title": "Title"}, Event: "Start"})
	file, err := disk.Save(filepath.Join(t.TempDir(), "run"), doc)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"contract", file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	var got struct {
		Design string
		Pages  []struct {
			Page   string
			Inputs []struct{ Name, Kind string }
			Events []struct{ Name string }
		}
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("%v\n%s", err, stdout.String())
	}
	if got.Design != "Run" || len(got.Pages) != 1 || len(got.Pages[0].Inputs) != 1 ||
		got.Pages[0].Inputs[0].Name != "Title" || got.Pages[0].Inputs[0].Kind != "string" ||
		len(got.Pages[0].Events) != 1 || got.Pages[0].Events[0].Name != "Start" {
		t.Fatalf("contract: %s", stdout.String())
	}
}
