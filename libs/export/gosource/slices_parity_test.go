package gosource

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// goRun runs a go command in dir and returns what it printed.
func goRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	cmd := exec.Command(goBin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return string(out)
}

// The slice layout draws what the flat layout draws, for every example design:
// the pages are the flat screens with their names exported, so the drawings
// are the same by construction, and this keeps it so.
func TestTheSliceLayoutDrawsWhatTheFlatLayoutDraws(t *testing.T) {
	if testing.Short() {
		t.Skip("builds projects with the go tool")
	}
	paths, err := filepath.Glob("../../../examples/*.cuppa")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no examples: %v", err)
	}
	for _, path := range paths {
		doc, err := disk.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		cat := cupp.Adopt(standard.Default(), doc.Embedded)
		ident := pascal(doc.Name)
		t.Run(filepath.Base(path), func(t *testing.T) {
			// Flat.
			flat := GenerateScreens([]design.Document{doc}, cat, "screens")
			flatDir := filepath.Join(t.TempDir(), "flat")
			if err := WriteScreens(flatDir, flat); err != nil {
				t.Fatal(err)
			}
			gomod := "module example.com/flat\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 " + bubblesVersion + "\n\tcharm.land/bubbletea/v2 " + bubbleteaVersion + "\n\tcharm.land/lipgloss/v2 " + lipglossVersion + "\n)\n"
			dump := fmt.Sprintf(`package screens

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDump(t *testing.T) {
	w, h := %sSize()
	fmt.Println("DRAW<<" + ansi.Strip(%s(Default%sProps(), w, h).View) + ">>DRAW")
}
`, ident, ident, ident)
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			must(os.WriteFile(filepath.Join(flatDir, "go.mod"), []byte(gomod), 0o644))
			must(os.WriteFile(filepath.Join(flatDir, "dump_test.go"), []byte(dump), 0o644))
			goRun(t, flatDir, "mod", "tidy")
			flatOut := between(goRun(t, flatDir, "test", "-run", "TestDump", "-v", "."))

			// Slices.
			p, err := GenerateSlices(doc, cat, Options{Module: "example.com/sliced"})
			if err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(t.TempDir(), "sliced")
			must(Write(dir, p))
			check := `package app

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDump(t *testing.T) {
	w, h := Size()
	fmt.Println("DRAW<<" + ansi.Strip(View(Default(), w, h).View) + ">>DRAW")
}
`
			must(os.WriteFile(filepath.Join(dir, "app", "dump_test.go"), []byte(check), 0o644))
			goRun(t, dir, "mod", "tidy")
			goRun(t, dir, "vet", "./...")
			slicedOut := between(goRun(t, dir, "test", "-run", "TestDump", "-v", "./app"))
			if flatOut == "" || flatOut != slicedOut {
				t.Errorf("the layouts draw differently:\n-- flat\n%s\n-- slices\n%s", flatOut, slicedOut)
			}
		})
	}
}

func between(out string) string {
	_, rest, ok := strings.Cut(out, "DRAW<<")
	if !ok {
		return ""
	}
	drawn, _, _ := strings.Cut(rest, ">>DRAW")
	return drawn
}
