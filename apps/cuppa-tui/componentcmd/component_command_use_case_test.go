package componentcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const good = `{
  "cuppa": 1, "name": "demo", "title": "Demo", "description": "A demo.", "license": "MIT",
  "go": { "module": "example.com/demo", "import": "example.com/demo", "package": "demo", "bubbletea": 2, "model": "Model", "constructor": "New" },
  "size": { "default": { "w": 10, "h": 3 }, "min": { "w": 4, "h": 1 } },
  "props": [ { "key": "text", "label": "Text", "kind": "text", "default": "hi", "go": { "option": "WithText" } } ]
}`

func TestLocationUnderstandsFilesAddressesAndRepositories(t *testing.T) {
	file := filepath.Join(t.TempDir(), "c.json")
	if err := os.WriteFile(file, []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		arg, want string
		isFile    bool
	}{
		{file, file, true},
		{"github.com/meta-ui-labs/bubble-colourpicker", "https://raw.githubusercontent.com/meta-ui-labs/bubble-colourpicker/HEAD/cuppa.component.json", false},
		{"github.com/meta-ui-labs/bubble-colourpicker@v0.1.1", "https://raw.githubusercontent.com/meta-ui-labs/bubble-colourpicker/v0.1.1/cuppa.component.json", false},
		{"https://github.com/meta-ui-labs/bubble-colourpicker/", "https://raw.githubusercontent.com/meta-ui-labs/bubble-colourpicker/HEAD/cuppa.component.json", false},
		{"https://example.com/x/cuppa.component.json", "https://example.com/x/cuppa.component.json", false},
	}
	for _, c := range cases {
		got, isFile := Location(c.arg)
		if got != c.want || isFile != c.isFile {
			t.Errorf("Location(%q) = %q, %v; want %q, %v", c.arg, got, isFile, c.want, c.isFile)
		}
	}
}

func TestCheckPrintsWhatTheDescriptionDescribes(t *testing.T) {
	file := filepath.Join(t.TempDir(), "c.json")
	if err := os.WriteFile(file, []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if code := Run([]string{"check", file}, &out, &errs); code != 0 {
		t.Fatalf("exit %d: %s%s", code, out.String(), errs.String())
	}
	for _, want := range []string{"community.demo", "import example.com/demo", "text", "the description is in order"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestCheckReportsProblemsAndFailsTheExitCode(t *testing.T) {
	file := filepath.Join(t.TempDir(), "c.json")
	if err := os.WriteFile(file, []byte(strings.Replace(good, `"license": "MIT"`, `"license": ""`, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if code := Run([]string{"check", file}, &out, &errs); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(out.String(), "license is empty") {
		t.Errorf("output:\n%s", out.String())
	}
	if code := Run([]string{"nonsense"}, &out, &errs); code != 2 {
		t.Errorf("usage exit %d, want 2", code)
	}
}

func TestChangeAndOptionsWorkOnADesignFile(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 80, 24))
	id, err := ed.Add("bubbles.textinput", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	ed.Rename(id, "Colour field")
	if err := ed.SetBinding(id, "value", "Colour"); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "d.cuppa")
	if _, err := disk.Save(file, ed.Document()); err != nil {
		t.Fatal(err)
	}

	var out, errs bytes.Buffer
	if code := Run([]string{"options", file, "colour field"}, &out, &errs); code != 0 || !strings.Contains(out.String(), "lipgloss.colourpicker") {
		t.Fatalf("options: exit %d\n%s%s", code, out.String(), errs.String())
	}
	out.Reset()
	if code := Run([]string{"change", file, "Colour field", "lipgloss.colourpicker"}, &out, &errs); code != 0 {
		t.Fatalf("change: exit %d\n%s%s", code, out.String(), errs.String())
	}
	doc, err := disk.Load(file)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := doc.Get(id)
	if n.Component != "lipgloss.colourpicker" || n.Bind["value"] != "Colour" {
		t.Errorf("after the change: %q %v", n.Component, n.Bind)
	}

	// A change that would lose a variable is refused, and the file is left alone.
	ed2 := editor.New(cat, doc)
	list, _ := ed2.Add("lipgloss.list", 2, 12)
	if err := ed2.SetBinding(list, "items", "Teas"); err != nil {
		t.Fatal(err)
	}
	if _, err := disk.Save(file, ed2.Document()); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errs.Reset()
	if code := Run([]string{"change", file, string(list), "lipgloss.label"}, &out, &errs); code != 1 || !strings.Contains(errs.String(), "Teas") {
		t.Fatalf("a lossy change: exit %d\n%s%s", code, out.String(), errs.String())
	}
	if after, _ := disk.Load(file); func() bool { n, _ := after.Get(list); return n.Component != "lipgloss.list" }() {
		t.Error("a refused change modified the file")
	}
	if code := Run([]string{"change", file, string(list), "lipgloss.label", "--allow-loss"}, &out, &errs); code != 0 {
		t.Errorf("allowed: exit %d\n%s", code, errs.String())
	}
}
