// Package filedialog is the Open / Save box: a folder browser you operate by
// mouse, with a file name field you can type into.
package filedialog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Spec says what the dialog is for.
type Spec struct {
	Title string
	// Dir is the folder to start in; empty means the working directory.
	Dir string
	// Name pre-fills the file name field.
	Name string
	// Ext is the extension (with dot) shown and appended; empty shows every file.
	Ext string
	// Also lists other extensions shown (and opened) besides Ext, such as the
	// Go source of an application that carries its design.
	Also []string
	// Save asks for a name to write to; otherwise an existing file is picked.
	Save bool
}

const (
	boxW      = 60
	listRows  = 12
	chromeRow = 6 // rows around the list inside the box
)

type entry struct {
	name string
	dir  bool
}

// Model is one open file dialog.
type Model struct {
	spec    Spec
	dir     string
	entries []entry
	scroll  int
	// picked is the entry highlighted by a click; a second click accepts it.
	picked int
	name   []rune
	errMsg string
	hover  int // list row under the pointer, -1 for none
	// typed is true once the name field has been typed into, so Enter accepts
	// the name rather than opening the highlighted folder.
	typed bool

	rect    design.Rect
	okSpan  span
	noSpan  span
	outcome modal.Outcome
	done    bool
}

type span struct{ x0, x1 int }

// New opens the dialog.
func New(spec Spec) *Model {
	dir := spec.Dir
	if dir == "" {
		dir, _ = os.Getwd()
	}
	m := &Model{spec: spec, name: []rune(spec.Name), picked: -1, hover: -1}
	m.cd(dir)
	return m
}

// Dir is the folder currently shown.
func (m *Model) Dir() string { return m.dir }

// cd lists dir; on failure it stays where it was and shows why.
func (m *Model) cd(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		m.errMsg = err.Error()
		return
	}
	items, err := os.ReadDir(abs)
	if err != nil {
		m.errMsg = "cannot open " + abs
		return
	}
	var dirs, files []entry
	for _, it := range items {
		if strings.HasPrefix(it.Name(), ".") {
			continue
		}
		switch {
		case it.IsDir():
			dirs = append(dirs, entry{it.Name(), true})
		case m.spec.Ext == "" || strings.EqualFold(filepath.Ext(it.Name()), m.spec.Ext) || m.also(filepath.Ext(it.Name())):
			files = append(files, entry{it.Name(), false})
		}
	}
	byName := func(s []entry) {
		sort.Slice(s, func(i, j int) bool { return strings.ToLower(s[i].name) < strings.ToLower(s[j].name) })
	}
	byName(dirs)
	byName(files)
	m.dir = abs
	m.entries = nil
	if filepath.Dir(abs) != abs {
		m.entries = append(m.entries, entry{"..", true})
	}
	m.entries = append(append(m.entries, dirs...), files...)
	m.scroll, m.picked, m.errMsg = 0, -1, ""
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) {
	w, h := min(boxW, sw), listRows+chromeRow
	m.rect = design.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/3, 0), W: w, H: h}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { return m.rect }

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	inner := m.rect.W - 2
	content := []string{theme.Dim(" " + ansi.Truncate(m.dir, inner-2, "…"))}
	for row := 0; row < listRows; row++ {
		content = append(content, m.listRow(row, inner))
	}
	field := string(m.name) + theme.Title("█")
	label := " Name: "
	content = append(content, theme.Fit(label+field, inner))
	status := ""
	if m.errMsg != "" {
		status = theme.Fit(" "+lipWarn(m.errMsg), inner)
	}
	content = append(content, status)
	ok := "Open"
	if m.spec.Save {
		ok = "Save"
	}
	content = append(content, m.buttons(ok, inner))
	return theme.Panel(m.spec.Title, content, m.rect.W)
}

func lipWarn(s string) string { return theme.Dim(s) }

func (m *Model) listRow(row, inner int) string {
	i := m.scroll + row
	if i >= len(m.entries) {
		return ""
	}
	e := m.entries[i]
	text := "  " + e.name
	if e.dir {
		text = "▸ " + e.name + "/"
	}
	text = theme.Fit(" "+text, inner)
	switch {
	case i == m.picked:
		return theme.Selected(text)
	case row == m.hover:
		return theme.Hovered(text)
	}
	return text
}

func (m *Model) buttons(ok string, inner int) string {
	a, b := "[ "+ok+" ]", "[ Cancel ]"
	x := max(inner-ansi.StringWidth(a)-ansi.StringWidth(b)-2, 1)
	m.okSpan = span{1 + x, 1 + x + ansi.StringWidth(a)}
	m.noSpan = span{m.okSpan.x1 + 1, m.okSpan.x1 + 1 + ansi.StringWidth(b)}
	return strings.Repeat(" ", x) + theme.Button(a, true) + " " + theme.Button(b, true)
}

// rowAt maps a box-relative y to a list row, or -1.
func rowAt(y int) int {
	if y >= 2 && y < 2+listRows {
		return y - 2
	}
	return -1
}

// Handle implements modal.Modal.
func (m *Model) Handle(e pointer.Event) {
	if m.done {
		return
	}
	x, y := e.X-m.rect.X, e.Y-m.rect.Y
	inBox := x >= 0 && y >= 0 && x < m.rect.W && y < m.rect.H
	switch e.Phase {
	case pointer.Wheel:
		m.scrollBy(e.WheelY * 3)
	case pointer.Move:
		m.hover = -1
		if inBox && x > 0 && x < m.rect.W-1 {
			if r := rowAt(y); m.scroll+r < len(m.entries) {
				m.hover = r
			}
		}
	case pointer.Down:
		if !e.Left || !inBox {
			return
		}
		switch {
		case y == m.rect.H-2 && x >= m.okSpan.x0 && x < m.okSpan.x1:
			m.accept()
		case y == m.rect.H-2 && x >= m.noSpan.x0 && x < m.noSpan.x1:
			m.done, m.outcome = true, modal.Outcome{Canceled: true}
		case rowAt(y) >= 0 && x > 0 && x < m.rect.W-1:
			m.clickEntry(m.scroll + rowAt(y))
		}
	}
}

func (m *Model) scrollBy(n int) {
	m.scroll = max(min(m.scroll+n, len(m.entries)-listRows), 0)
}

// clickEntry enters folders; for files the first click picks and the second
// accepts (so Open is two clicks on the same file, or one click then Open).
func (m *Model) clickEntry(i int) {
	if i >= len(m.entries) {
		return
	}
	e := m.entries[i]
	if e.dir {
		m.cd(filepath.Join(m.dir, e.name))
		return
	}
	if m.picked == i && !m.spec.Save {
		m.accept()
		return
	}
	m.picked = i
	m.name = []rune(e.name)
}

// Key implements modal.Modal.
func (m *Model) Key(text string, back, enter, esc bool) {
	switch {
	case esc:
		m.done, m.outcome = true, modal.Outcome{Canceled: true}
	case enter && m.picked >= 0 && m.picked < len(m.entries) && m.entries[m.picked].dir && !m.typed:
		m.clickEntry(m.picked)
	case enter:
		m.accept()
	case back:
		m.typed = true
		if len(m.name) > 0 {
			m.name = m.name[:len(m.name)-1]
		}
	case text != "":
		m.name = append(m.name, []rune(text)...)
		m.picked, m.typed = -1, true
	}
}

// Nav implements modal.Navigator: Up and Down move the highlight through the
// list (a file's name goes into the name field), PageUp and PageDown move by a
// page, Home and End jump, and Alt+Up goes to the parent folder. Enter opens a
// highlighted folder, or accepts the name.
func (m *Model) Nav(name string) bool {
	last := len(m.entries) - 1
	switch name {
	case "up":
		m.highlight(m.picked - 1)
	case "down":
		m.highlight(m.picked + 1)
	case "pgup":
		m.highlight(m.picked - listRows)
	case "pgdown":
		m.highlight(m.picked + listRows)
	case "home":
		m.highlight(0)
	case "end":
		m.highlight(last)
	case "alt+up":
		m.cd(filepath.Dir(m.dir))
	default:
		return false
	}
	return true
}

// highlight moves the highlight to entry i (kept in range) and scrolls to it.
func (m *Model) highlight(i int) {
	if len(m.entries) == 0 {
		return
	}
	i = min(max(i, 0), len(m.entries)-1)
	m.picked, m.typed = i, false
	switch {
	case i < m.scroll:
		m.scroll = i
	case i >= m.scroll+listRows:
		m.scroll = i - listRows + 1
	}
	if e := m.entries[i]; !e.dir {
		m.name = []rune(e.name)
	}
}

// accept resolves the name field: a folder navigates, a file ends the dialog.
func (m *Model) accept() {
	name := strings.TrimSpace(string(m.name))
	if name == "" {
		m.errMsg = "type a file name"
		return
	}
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.dir, name)
	}
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		m.cd(path)
		m.name = nil
		return
	}
	if m.spec.Ext != "" && !strings.EqualFold(filepath.Ext(path), m.spec.Ext) {
		if m.spec.Save {
			path += m.spec.Ext
		}
	}
	if !m.spec.Save {
		if fi, err := os.Stat(path); err != nil || fi.IsDir() {
			m.errMsg = "no such file: " + name
			return
		}
	}
	m.done, m.outcome = true, modal.Outcome{Path: path}
}

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }

// Describe reads the dialog out: the folder, its entries, the file name and buttons.
func (m *Model) Describe() []a11y.Node {
	var items []a11y.Node
	for i, e := range m.entries {
		label := e.name
		if e.dir {
			label += ", folder"
		}
		items = append(items, a11y.Item(label, i == m.picked))
	}
	name := a11y.Field("File name", string(m.name))
	name.Focused = true
	nodes := []a11y.Node{a11y.Heading(m.spec.Title), a11y.Text("Folder " + m.dir), a11y.List("Files", items...), name}
	if m.errMsg != "" {
		nodes = append(nodes, a11y.Text(m.errMsg))
	}
	return append(nodes, a11y.Button("OK"), a11y.Button("Cancel"))
}

// also reports whether ext is one of the extra extensions the dialog shows.
func (m *Model) also(ext string) bool {
	for _, e := range m.spec.Also {
		if strings.EqualFold(e, ext) {
			return true
		}
	}
	return false
}
