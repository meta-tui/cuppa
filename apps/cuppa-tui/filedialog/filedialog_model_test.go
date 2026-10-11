package filedialog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
)

func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"b.cuppa", "a.cuppa", "notes.txt", ".hidden.cuppa"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func click(m *Model, listRow int) {
	m.Handle(pointer.Event{X: m.rect.X + 5, Y: m.rect.Y + 2 + listRow, Phase: pointer.Down, Left: true})
}

func names(m *Model) []string {
	var out []string
	for _, e := range m.entries {
		out = append(out, e.name)
	}
	return out
}

func TestListsFoldersFirstAndFiltersByExtension(t *testing.T) {
	m := New(Spec{Title: "Open", Dir: fixture(t), Ext: ".cuppa"})
	got := strings.Join(names(m), ",")
	if got != "..,sub,a.cuppa,b.cuppa" {
		t.Fatalf("got %s", got)
	}
}

func TestClickFileFillsNameAndSecondClickOpens(t *testing.T) {
	dir := fixture(t)
	m := New(Spec{Title: "Open", Dir: dir, Ext: ".cuppa"})
	m.Place(100, 30)
	click(m, 2) // a.cuppa
	if string(m.name) != "a.cuppa" {
		t.Fatalf("name = %q", string(m.name))
	}
	if _, done := m.Outcome(); done {
		t.Fatal("one click should only pick")
	}
	click(m, 2)
	out, done := m.Outcome()
	if !done || out.Path != filepath.Join(dir, "a.cuppa") {
		t.Fatalf("got %+v", out)
	}
}

func TestClickingAFolderEntersIt(t *testing.T) {
	dir := fixture(t)
	m := New(Spec{Title: "Open", Dir: dir, Ext: ".cuppa"})
	m.Place(100, 30)
	click(m, 1) // sub
	if m.Dir() != filepath.Join(dir, "sub") {
		t.Fatalf("dir = %s", m.Dir())
	}
	click(m, 0) // ..
	if m.Dir() != dir {
		t.Fatalf("dir = %s", m.Dir())
	}
}

func TestSaveAddsExtensionAndAllowsNewNames(t *testing.T) {
	dir := fixture(t)
	m := New(Spec{Title: "Save", Dir: dir, Ext: ".cuppa", Save: true})
	for _, r := range "fresh" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	out, done := m.Outcome()
	if !done || out.Path != filepath.Join(dir, "fresh.cuppa") {
		t.Fatalf("got %+v", out)
	}
}

func TestOpenRejectsMissingFile(t *testing.T) {
	m := New(Spec{Title: "Open", Dir: fixture(t), Ext: ".cuppa"})
	m.Key("n", false, false, false)
	m.Key("", false, true, false)
	if _, done := m.Outcome(); done || m.errMsg == "" {
		t.Fatalf("done=%v err=%q", done, m.errMsg)
	}
}

func TestTypingAFolderPathNavigates(t *testing.T) {
	dir := fixture(t)
	m := New(Spec{Title: "Open", Dir: dir, Ext: ".cuppa"})
	for _, r := range "sub" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	if m.Dir() != filepath.Join(dir, "sub") || len(m.name) != 0 {
		t.Fatalf("dir=%s name=%q", m.Dir(), string(m.name))
	}
}

func TestBackspaceEscAndCancelButton(t *testing.T) {
	m := New(Spec{Title: "Open", Dir: fixture(t), Name: "ab"})
	m.Key("", true, false, false)
	if string(m.name) != "a" {
		t.Fatalf("name = %q", string(m.name))
	}
	m.Key("", false, false, true)
	if out, done := m.Outcome(); !done || !out.Canceled {
		t.Fatalf("esc: %+v", out)
	}
	m = New(Spec{Title: "Open", Dir: fixture(t)})
	m.Place(100, 30)
	m.Lines()
	m.Handle(pointer.Event{X: m.rect.X + m.noSpan.x0 + 1, Y: m.rect.Y + m.rect.H - 2, Phase: pointer.Down, Left: true})
	if out, done := m.Outcome(); !done || !out.Canceled {
		t.Fatalf("cancel button: %+v", out)
	}
}

func TestRendersExactlyItsRect(t *testing.T) {
	m := New(Spec{Title: "Open", Dir: fixture(t), Ext: ".cuppa"})
	m.Place(100, 30)
	lines := m.Lines()
	if len(lines) != m.rect.H {
		t.Fatalf("lines %d rect %d", len(lines), m.rect.H)
	}
	for _, l := range lines {
		if ansi.StringWidth(l) != m.rect.W {
			t.Fatalf("width %d: %q", ansi.StringWidth(l), ansi.Strip(l))
		}
	}
}

func TestWheelScrollsLongListsWithinBounds(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 30; i++ {
		_ = os.WriteFile(filepath.Join(dir, strings.Repeat("f", i+1)+".cuppa"), nil, 0o644)
	}
	m := New(Spec{Title: "Open", Dir: dir, Ext: ".cuppa"})
	m.Place(100, 30)
	for i := 0; i < 20; i++ {
		m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1})
	}
	if m.scroll != len(m.entries)-listRows {
		t.Fatalf("scroll = %d", m.scroll)
	}
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: -1})
	m.scrollBy(-100)
	if m.scroll != 0 {
		t.Fatalf("scroll = %d", m.scroll)
	}
}
