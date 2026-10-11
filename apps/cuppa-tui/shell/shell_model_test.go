package shell

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

func newShell(t *testing.T) *Model {
	t.Helper()
	m := New(standard.Default())
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	return m
}

var errNoFreeze = errors.New("no freeze")

func send(m *Model, msg tea.Msg) { m.Update(msg) }

func click(x, y int) tea.Msg  { return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft} }
func motion(x, y int) tea.Msg { return tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft} }

// motion2 is the pointer moving with no button held.
func motion2(x, y int) tea.Msg { return tea.MouseMotionMsg{X: x, Y: y} }
func release(x, y int) tea.Msg { return tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft} }

func TestLayoutFillsTheTerminal(t *testing.T) {
	m := newShell(t)
	out := strings.Split(m.render(), "\n")
	if len(out) != 40 {
		t.Fatalf("rows = %d", len(out))
	}
}

func TestDragFromPaletteAndDropCreatesNode(t *testing.T) {
	m := newShell(t)
	// Palette: line 0 is the shell title bar, so pane row 2 is screen row 3
	// (first family header); the first item sits on screen row 4.
	send(m, click(5, firstComponentY))
	if m.dragging == "" {
		t.Fatal("pressing a palette item should start a drag")
	}
	stageX := m.layout.stage.X + 10
	send(m, motion(stageX, 8))
	if m.stg == nil || m.dragging == "" {
		t.Fatal("drag lost")
	}
	send(m, release(stageX, 8))
	doc := m.Editor().Document()
	if len(doc.Nodes) != 1 {
		t.Fatalf("nodes = %d", len(doc.Nodes))
	}
	if r := doc.Nodes[0].Rect; r.X != 10 || r.Y != 8-m.layout.stage.Y {
		t.Fatalf("dropped at %+v, want cell (10,%d)", r, 8-m.layout.stage.Y)
	}
	if m.dragging != "" {
		t.Fatal("drag state not cleared")
	}
}

func TestReleasingOutsideTheCanvasCancelsTheDrag(t *testing.T) {
	m := newShell(t)
	send(m, click(5, firstComponentY))
	send(m, motion(m.layout.stage.X+5, 8))
	send(m, release(2, 10)) // back over the palette
	if n := len(m.Editor().Document().Nodes); n != 0 {
		t.Fatalf("nodes = %d", n)
	}
	if m.dragging != "" {
		t.Fatal("drag state not cleared")
	}
}

func TestDroppedNodeCanBeSelectedMovedAndInspected(t *testing.T) {
	m := newShell(t)
	send(m, click(5, firstComponentY))
	x0 := m.layout.stage.X
	send(m, motion(x0+10, 8))
	send(m, release(x0+10, 8))
	// Select by clicking inside, then drag 6 cells right.
	send(m, click(x0+12, 9))
	send(m, motion(x0+18, 9))
	send(m, release(x0+18, 9))
	n, ok := m.Editor().Primary()
	if !ok || n.Rect.X != 16 {
		t.Fatalf("node = %+v ok=%v", n, ok)
	}
	if !strings.Contains(m.render(), n.Name) {
		t.Fatal("inspector/layers should mention the selected node")
	}
}

func TestKeyboardBasics(t *testing.T) {
	m := newShell(t)
	send(m, click(5, firstComponentY))
	send(m, release(m.layout.stage.X+3, 5))
	if len(m.Editor().Document().Nodes) != 1 {
		t.Fatal("setup failed")
	}
	send(m, tea.KeyPressMsg{Code: tea.KeyDelete})
	if len(m.Editor().Document().Nodes) != 0 {
		t.Fatal("delete key should remove the selection")
	}
	// The design changed since it was loaded, so quitting asks first.
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}); cmd != nil {
		t.Fatal("ctrl+q on unsaved work must ask, not quit")
	}
	if m.flow.Modal() == nil {
		t.Fatal("expected the unsaved-changes dialog")
	}
	clickText(t, m, "Discard")
	if !m.flow.Quitting() {
		t.Fatal("discarding should quit")
	}
}

func TestCtrlQQuitsAtOnceWhenNothingIsUnsaved(t *testing.T) {
	m := newShell(t)
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}); cmd == nil {
		t.Fatal("clean design should quit")
	}
}

// screenText is the rendered screen without colour codes.
func screenText(m *Model) []string {
	lines := strings.Split(m.render(), "\n")
	for i, l := range lines {
		lines[i] = ansi.Strip(l)
	}
	return lines
}

// clickText clicks the first screen cell where text appears.
func clickText(t *testing.T, m *Model, text string) {
	t.Helper()
	for y, l := range screenText(m) {
		if i := strings.Index(l, text); i >= 0 {
			x := len([]rune(l[:i])) + 1
			send(m, click(x, y))
			send(m, release(x, y))
			return
		}
	}
	t.Fatalf("%q not on screen:\n%s", text, strings.Join(screenText(m), "\n"))
}

func TestMenuBarIsOnTopAndDropdownsOverlayThePanes(t *testing.T) {
	m := newShell(t)
	top := screenText(m)[0]
	for _, label := range []string{"File", "Edit", "Export", "Help"} {
		if !strings.Contains(top, label) {
			t.Fatalf("menu bar lacks %s: %q", label, top)
		}
	}
	clickText(t, m, "Export")
	screen := strings.Join(screenText(m), "\n")
	if !strings.Contains(screen, "Image (PNG)") {
		t.Fatalf("dropdown not shown:\n%s", screen)
	}
	if len(strings.Split(screen, "\n")) != 40 {
		t.Fatal("dropdown must not change the screen height")
	}
	send(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if strings.Contains(strings.Join(screenText(m), "\n"), "Image (PNG)") {
		t.Fatal("Esc should close the dropdown")
	}
}

func TestFileMenuSaveAsOpensDialogAndSaves(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	m := newShell(t)
	send(m, click(5, firstComponentY))
	send(m, release(m.layout.stage.X+3, 5)) // drop a component so there is something to save
	clickText(t, m, "File")
	clickText(t, m, "Save As")
	if m.flow.Modal() == nil {
		t.Fatal("Save As should open a dialog")
	}
	if !strings.Contains(strings.Join(screenText(m), "\n"), "Save design") {
		t.Fatal("dialog not drawn")
	}
	send(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.flow.Modal() != nil || m.Editor().Dirty() {
		t.Fatalf("not saved: modal=%v dirty=%v", m.flow.Modal() != nil, m.Editor().Dirty())
	}
	if !strings.HasSuffix(m.flow.Path(), "Untitled.cuppa") {
		t.Fatalf("path = %q", m.flow.Path())
	}
	if !strings.Contains(screenText(m)[0], "Untitled.cuppa") {
		t.Fatal("title should show the file name")
	}
}

func TestDialogCapturesTheMouse(t *testing.T) {
	m := newShell(t)
	m.flow.Notice("Hello", "world")
	before := len(m.Editor().Document().Nodes)
	send(m, click(5, firstComponentY)) // would start a palette drag if the dialog let it through
	if m.dragging != "" || len(m.Editor().Document().Nodes) != before {
		t.Fatal("clicks must not reach the panes behind a dialog")
	}
}

func TestNarrowTerminalStillHasAStage(t *testing.T) {
	l := computeLayout(60, 20, 0, 0)
	if l.stage.W < 1 || l.palette.W < 16 || l.inspector.W < 20 {
		t.Fatalf("layout = %+v", l)
	}
}

func TestImageExportsAreGreyedOutWithoutFreezeAndExplainWhenClicked(t *testing.T) {
	m := newShell(t)
	m.flow.SetFreezeLookup(func() (string, error) { return "", errNoFreeze })
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m.settle()
	clickText(t, m, "Export")
	clickText(t, m, "Image (PNG)")
	dlg := m.flow.Modal()
	if dlg == nil {
		t.Fatal("clicking a greyed-out image export should explain")
	}
	if !strings.Contains(strings.Join(screenText(m), " "), "Freeze is needed") {
		t.Fatal("popup not drawn")
	}
	// Text export is unaffected.
	send(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	clickText(t, m, "Export")
	clickText(t, m, "Plain text")
	if m.flow.Modal() == nil || strings.Contains(strings.Join(screenText(m), " "), "Freeze is needed") {
		t.Fatal("plain text export should open the file dialog")
	}
}

func TestOpenFileOrNotifyShowsADialogWhenTheFileIsBad(t *testing.T) {
	m := newShell(t)
	m.OpenFileOrNotify(t.TempDir() + "/missing.cuppa")
	if m.flow.Modal() == nil {
		t.Fatal("a file that cannot be opened should be explained in a dialog")
	}
	if !strings.Contains(strings.Join(screenText(m), " "), "Cannot open file") {
		t.Fatal("dialog not drawn")
	}
}

// The pages bar sits under the canvas; Alt+PageUp and Alt+PageDown walk pages.
func TestThePagesBarIsUnderTheCanvasAndAltPageKeysWalkPages(t *testing.T) {
	m := newShell(t)
	if m.layout.pages.H != 1 || m.layout.pages.Y != m.layout.stage.Y+m.layout.stage.H {
		t.Fatalf("layout = %+v", m.layout)
	}
	if err := m.Editor().AddPage("Second"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.render(), "Second  2/2") {
		t.Fatal("the bar names the page being edited")
	}
	send(m, key(tea.KeyPgUp, tea.ModAlt))
	if m.Editor().ActivePage() != 0 {
		t.Fatalf("active = %d", m.Editor().ActivePage())
	}
	send(m, key(tea.KeyPgDown, tea.ModAlt))
	if m.Editor().ActivePage() != 1 {
		t.Fatalf("active = %d", m.Editor().ActivePage())
	}
}
