package componentcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
