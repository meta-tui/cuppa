// Package fileflow runs the file-related user journeys: New, Open, Save,
// Save As, Quit and the exports. Each one is a short chain of dialogs ending
// in a call to the file libraries. It owns the current file path.
package fileflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/confirm"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/filedialog"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/logo"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/image"
	"github.com/meta-ui-labs/cuppa/libs/export/text"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

// Default canvas of a new design, in cells.
const (
	defaultWidth  = 120
	defaultHeight = 40
)

// Job is slow work the shell runs off the UI thread, then reports back with
// Flow.Finish.
type Job struct {
	Run func() error
}

// Flow holds the state shared by every journey.
type Flow struct {
	ed  *editor.Editor
	cat scene.Catalog
	// onLoad is told about every loaded design.
	onLoad func(design.Document)

	path   string
	dir    string
	status string
	quit   bool

	modal  modal.Modal
	onDone func(modal.Outcome)
	sw, sh int

	job      *Job
	jobLabel string

	// locate finds freeze; tests replace it.
	locate func() (string, error)
	// freeze caches whether freeze was found at the last check.
	freeze bool
}

// New returns a flow editing ed.
func New(ed *editor.Editor, cat scene.Catalog) *Flow {
	f := &Flow{ed: ed, cat: cat, locate: image.Locate}
	f.CheckFreeze()
	return f
}

// Path is the file the design was last loaded from or saved to, or "".
func (f *Flow) Path() string { return f.path }

// Title is the name to show for the design: the file name, else the design name.
func (f *Flow) Title() string {
	if f.path != "" {
		return filepath.Base(f.path)
	}
	return f.ed.Document().Name
}

// Status is the last result worth telling the user ("Saved demo.cuppa").
func (f *Flow) Status() string { return f.status }

// SetStatus replaces the feedback line, for modes that explain themselves.
func (f *Flow) SetStatus(s string) { f.status = s }

// Quitting is true once the user confirmed leaving.
func (f *Flow) Quitting() bool { return f.quit }

// Modal is the dialog currently open, or nil.
func (f *Flow) Modal() modal.Modal { return f.modal }

// SetScreen tells the flow how big the screen is, to centre dialogs.
func (f *Flow) SetScreen(w, h int) {
	f.sw, f.sh = w, h
	if f.modal != nil {
		f.modal.Place(w, h)
	}
}

// Resolve advances the journey once the open dialog has ended. Call it after
// every input that was routed to the dialog.
func (f *Flow) Resolve() {
	if f.modal == nil {
		return
	}
	out, done := f.modal.Outcome()
	if !done {
		return
	}
	cb := f.onDone
	f.modal, f.onDone = nil, nil
	cb(out)
}

// TakeJob hands over pending slow work, once.
func (f *Flow) TakeJob() *Job {
	j := f.job
	f.job = nil
	return j
}

// Finish reports how the job ended.
func (f *Flow) Finish(err error) {
	if err != nil {
		f.status = "Export failed"
		f.Notice("Export failed", err.Error())
		return
	}
	f.status = f.jobLabel
}

// Notice shows a message with an OK button.
func (f *Flow) Notice(title, body string) {
	f.show(confirm.New(title, body, "OK"), func(modal.Outcome) {})
}

// NoticeWithArt is Notice with a picture above the message, such as the logo.
func (f *Flow) NoticeWithArt(title, body string, art []string) {
	f.show(confirm.New(title, body, "OK").WithArt(art), func(modal.Outcome) {})
}

// Show opens any dialog (for example a colour picker) and calls cb when it ends.
func (f *Flow) Show(m modal.Modal, cb func(modal.Outcome)) { f.show(m, cb) }

func (f *Flow) show(m modal.Modal, cb func(modal.Outcome)) {
	m.Place(f.sw, f.sh)
	f.modal, f.onDone = m, cb
}

// SetOnLoad sets a function told about every design that was just loaded (new, opened
// or from the command line), so the app can adopt the components it embeds.
func (f *Flow) SetOnLoad(fn func(design.Document)) { f.onLoad = fn }

func (f *Flow) load(doc design.Document) {
	f.ed.Load(doc)
	if f.onLoad != nil {
		f.onLoad(doc)
	}
}

// NewDesign starts an empty design, offering to save the current one first.
func (f *Flow) NewDesign() {
	f.guard(func() {
		f.load(design.NewDocument("Untitled", defaultWidth, defaultHeight))
		f.path, f.status = "", "New design"
	})
}

// Quit leaves the app, offering to save first.
func (f *Flow) Quit() { f.guard(func() { f.quit = true }) }

// Open asks for a file and loads it.
func (f *Flow) Open() {
	f.guard(func() {
		d := filedialog.New(filedialog.Spec{Title: "Open design", Dir: f.dir, Ext: format.Extension, Also: []string{".go"}})
		f.show(d, func(o modal.Outcome) {
			if o.Canceled {
				return
			}
			doc, err := disk.Load(o.Path)
			if err != nil {
				f.Notice("Cannot open file", describe(err))
				return
			}
			f.load(doc)
			f.path, f.dir = o.Path, filepath.Dir(o.Path)
			f.status = "Opened " + filepath.Base(o.Path)
			if disk.IsGoSource(o.Path) {
				// The design came out of Go source: there is no .cuppa to save
				// over, the first save asks for a name.
				f.path = ""
				f.status = "Opened the design in " + filepath.Base(o.Path) + "; save it as a .cuppa file"
			}
		})
	})
}

// OpenPath loads a file straight away, for a path given on the command line.
func (f *Flow) OpenPath(path string) error {
	doc, err := disk.Load(path)
	if err != nil {
		return fmt.Errorf("%s: %s", path, describe(err))
	}
	f.load(doc)
	f.path, f.dir = path, filepath.Dir(path)
	f.status = "Opened " + filepath.Base(path)
	if disk.IsGoSource(path) {
		f.path = ""
		f.status = "Opened the design in " + filepath.Base(path) + "; save it as a .cuppa file"
	}
	return nil
}

// Save writes to the current file, asking for a name the first time.
func (f *Flow) Save() { f.save(func() {}) }

func (f *Flow) save(then func()) {
	if f.path == "" {
		f.saveAs(then)
		return
	}
	if f.write(f.path) {
		then()
	}
}

// SaveAs asks for a name and writes there.
func (f *Flow) SaveAs() { f.saveAs(func() {}) }

func (f *Flow) saveAs(then func()) {
	name := strings.TrimSuffix(f.Title(), format.Extension)
	d := filedialog.New(filedialog.Spec{Title: "Save design", Dir: f.startDir(), Name: name, Ext: format.Extension, Save: true})
	f.show(d, func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		f.confirmOverwrite(o.Path, func() {
			if f.write(o.Path) {
				then()
			}
		})
	})
}

// confirmOverwrite runs ok straight away, or after the user agrees to replace
// an existing file (other than the one already open).
func (f *Flow) confirmOverwrite(path string, ok func()) {
	if _, err := os.Stat(path); err != nil || path == f.path {
		ok()
		return
	}
	body := fmt.Sprintf("%s already exists. Replace it?", filepath.Base(path))
	f.show(confirm.New("Replace file", body, "Replace", "Cancel"), func(o modal.Outcome) {
		if o.Button == "Replace" {
			ok()
		}
	})
}

func (f *Flow) write(path string) bool {
	saved, err := disk.Save(path, cupp.Embed(f.ed.Document(), f.cat))
	if err != nil {
		f.Notice("Cannot save file", describe(err))
		return false
	}
	f.ed.MarkSaved()
	f.path, f.dir = saved, filepath.Dir(saved)
	f.status = "Saved " + filepath.Base(saved)
	return true
}

// guard runs then now when nothing is unsaved, otherwise after the user chose
// to save or discard.
func (f *Flow) guard(then func()) {
	if !f.ed.Dirty() {
		then()
		return
	}
	body := fmt.Sprintf("Save changes to %s?", f.Title())
	f.show(confirm.New("Unsaved changes", body, "Save", "Discard", "Cancel"), func(o modal.Outcome) {
		switch o.Button {
		case "Save":
			f.save(then)
		case "Discard":
			then()
		}
	})
}

// ExportImage asks for a file name and renders the design through Freeze.
func (f *Flow) ExportImage(kind image.Format) {
	if !f.CheckFreeze() {
		f.Notice("Freeze is needed for images", freezeHelp)
		return
	}
	f.askExport("Export "+strings.ToUpper(string(kind)), "."+string(kind), func(path string) {
		doc := f.ed.Document()
		ansi := text.ANSI(doc, f.cat)
		opts := image.Options{Window: true}
		if c, ok := space.Resolve(doc.Background); ok {
			opts.Background = c.Hex() // the picture's backdrop matches the canvas
		}
		f.job = &Job{Run: func() error { return image.Write(path, ansi, opts) }}
		f.jobLabel = "Exported " + filepath.Base(path)
		f.status = "Exporting " + filepath.Base(path) + "…"
	})
}

// ExportText writes the design as coloured (ANSI) or plain text.
func (f *Flow) ExportText(colour bool) {
	ext, title := ".txt", "Export plain text"
	if colour {
		ext, title = ".ans", "Export colour text"
	}
	f.askExport(title, ext, func(path string) {
		doc := f.ed.Document()
		out := text.Plain(doc, f.cat)
		if colour {
			out = text.ANSI(doc, f.cat)
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			f.Notice("Cannot write file", describe(err))
			return
		}
		f.status = "Exported " + filepath.Base(path)
	})
}

func (f *Flow) askExport(title, ext string, do func(path string)) {
	name := strings.TrimSuffix(f.Title(), format.Extension)
	d := filedialog.New(filedialog.Spec{Title: title, Dir: f.startDir(), Name: name, Ext: ext, Save: true})
	f.show(d, func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		f.dir = filepath.Dir(o.Path)
		f.confirmOverwrite(o.Path, func() { do(o.Path) })
	})
}

func (f *Flow) startDir() string {
	if f.path != "" {
		return filepath.Dir(f.path)
	}
	return f.dir
}

// describe turns a file error into a sentence for the user.
func describe(err error) string {
	switch {
	case errors.Is(err, format.ErrNotCuppa):
		return "That file is not a Cuppa design."
	case errors.Is(err, format.ErrTooNew):
		return "That design was made with a newer version of Cuppa. Update Cuppa to open it."
	case errors.Is(err, format.ErrCorrupt):
		return "That design file is damaged and cannot be read."
	case errors.Is(err, os.ErrNotExist):
		return "That file does not exist."
	case errors.Is(err, os.ErrPermission):
		return "Permission denied."
	}
	return err.Error()
}

// freezeHelp tells the user how to get Freeze.
const freezeHelp = "Exporting PNG, SVG and WebP pictures uses Freeze, a separate tool from Charm that is not installed on this computer.\n\n" +
	"Install it with one of:\n" +
	"  go install github.com/charmbracelet/freeze@latest\n" +
	"  brew install charmbracelet/tap/freeze\n" +
	"  or download it from github.com/charmbracelet/freeze/releases\n\n" +
	"Then restart Cuppa. If it is installed somewhere unusual, set CUPPA_FREEZE to its path. Text export works without it."

// SetFreezeLookup replaces how freeze is found, for tests of front ends.
func (f *Flow) SetFreezeLookup(locate func() (string, error)) {
	f.locate = locate
	f.CheckFreeze()
}

// FreezeAvailable is whether freeze was found at the last check.
func (f *Flow) FreezeAvailable() bool { return f.freeze }

// CheckFreeze looks for freeze again and returns whether it was found.
func (f *Flow) CheckFreeze() bool {
	_, err := f.locate()
	f.freeze = err == nil
	return f.freeze
}

// Welcome tells a first-time user that picture export needs Freeze, once:
// marker is a file recording that the notice was shown. It does nothing when
// freeze is installed or the notice was shown before.
func (f *Flow) Welcome(marker string) {
	if f.freeze || marker == "" {
		return
	}
	if _, err := os.Stat(marker); err == nil {
		return
	}
	f.NoticeWithArt("Welcome to Cuppa", freezeHelp, logo.Dialog())
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err == nil {
		_ = os.WriteFile(marker, []byte("shown\n"), 0o644)
	}
}
