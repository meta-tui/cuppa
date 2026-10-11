package fileflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/image"
)

func setup(t *testing.T) (*Flow, *editor.Editor, string) {
	t.Helper()
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("Untitled", 60, 20))
	f := New(ed, cat)
	f.SetScreen(120, 40)
	dir := t.TempDir()
	f.dir = dir
	return f, ed, dir
}

func typeText(f *Flow, s string) {
	for _, r := range s {
		f.Modal().Key(string(r), false, false, false)
	}
}

func press(f *Flow, enter, esc bool) {
	f.Modal().Key("", false, enter, esc)
	f.Resolve()
}

func clickButton(t *testing.T, f *Flow, label string) {
	t.Helper()
	m := f.Modal()
	for y, l := range m.Lines() {
		if i := strings.Index(stripped(l), "[ "+label+" ]"); i >= 0 {
			r := m.Rect()
			m.Handle(pointer.Event{X: r.X + len([]rune(stripped(l)[:i])) + 2, Y: r.Y + y, Phase: pointer.Down, Left: true})
			f.Resolve()
			return
		}
	}
	t.Fatalf("no %q button", label)
}

func stripped(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			esc = true
		case esc && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'):
			esc = false
		case !esc:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TestSaveAsThenSaveWritesTheSameFile(t *testing.T) {
	f, ed, dir := setup(t)
	if _, err := ed.Add("lipgloss.box", 2, 2); err != nil {
		t.Fatal(err)
	}
	f.Save() // no path yet: asks for a name
	if f.Modal() == nil {
		t.Fatal("expected a Save dialog")
	}
	typeText(f, "") // name pre-filled "Untitled"
	press(f, true, false)
	want := filepath.Join(dir, "Untitled.cuppa")
	if f.Path() != want || ed.Dirty() {
		t.Fatalf("path=%q dirty=%v", f.Path(), ed.Dirty())
	}
	if _, err := ed.Add("huh.input", 5, 5); err != nil {
		t.Fatal(err)
	}
	f.Save() // straight to disk, no dialog
	if f.Modal() != nil || ed.Dirty() {
		t.Fatal("second save should not ask")
	}
	doc, err := disk.Load(want)
	if err != nil || len(doc.Nodes) != 2 {
		t.Fatalf("doc: %v %v", err, doc.Nodes)
	}
	if f.Status() != "Saved Untitled.cuppa" {
		t.Fatalf("status = %q", f.Status())
	}
}

func TestNewOnDirtyDesignOffersSaveDiscardCancel(t *testing.T) {
	f, ed, _ := setup(t)
	_, _ = ed.Add("lipgloss.box", 2, 2)
	f.NewDesign()
	clickButton(t, f, "Cancel")
	if len(ed.Document().Nodes) != 1 {
		t.Fatal("cancel must keep the design")
	}
	f.NewDesign()
	clickButton(t, f, "Discard")
	if len(ed.Document().Nodes) != 0 || f.Path() != "" {
		t.Fatalf("discard should start fresh: %+v", ed.Document().Nodes)
	}
}

func TestQuitOnDirtyDesignSavesThenQuits(t *testing.T) {
	f, ed, dir := setup(t)
	_, _ = ed.Add("lipgloss.box", 2, 2)
	f.Quit()
	if f.Quitting() {
		t.Fatal("quit must wait for the answer")
	}
	clickButton(t, f, "Save") // opens the Save As dialog
	typeText(f, "x")
	press(f, true, false)
	if !f.Quitting() {
		t.Fatal("should quit after saving")
	}
	if _, err := os.Stat(filepath.Join(dir, "Untitledx.cuppa")); err != nil {
		t.Fatal(err)
	}
}

func TestQuitOnCleanDesignQuitsAtOnce(t *testing.T) {
	f, _, _ := setup(t)
	f.Quit()
	if !f.Quitting() {
		t.Fatal("clean design should quit immediately")
	}
}

func TestOpenLoadsAFile(t *testing.T) {
	f, ed, dir := setup(t)
	doc := design.NewDocument("Other", 50, 10)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 4, H: 3}})
	if _, err := disk.Save(filepath.Join(dir, "other"), doc); err != nil {
		t.Fatal(err)
	}
	f.Open()
	typeText(f, "nope.cuppa")
	press(f, true, false)
	if f.Modal() == nil {
		t.Fatal("a missing file must keep the dialog open")
	}
	for range "nope.cuppa" {
		f.Modal().Key("", true, false, false)
	}
	typeText(f, "other.cuppa")
	press(f, true, false)
	if f.Modal() != nil {
		t.Fatalf("expected the dialog to finish; status=%q", f.Status())
	}
	if ed.Document().Width != 50 || f.Path() != filepath.Join(dir, "other.cuppa") || ed.Dirty() {
		t.Fatalf("not loaded: %+v path=%q", ed.Document(), f.Path())
	}
}

func TestOpenReportsDamagedFiles(t *testing.T) {
	f, _, dir := setup(t)
	bad := filepath.Join(dir, "bad.cuppa")
	_ = os.WriteFile(bad, []byte("not a design"), 0o644)
	f.Open()
	typeText(f, "bad.cuppa")
	press(f, true, false)
	m := f.Modal()
	if m == nil {
		t.Fatal("expected an error notice")
	}
	body := stripped(strings.Join(m.Lines(), "\n"))
	if !strings.Contains(body, "not a Cuppa design") {
		t.Fatalf("notice:\n%s", body)
	}
}

func TestSaveAsAsksBeforeReplacing(t *testing.T) {
	f, ed, dir := setup(t)
	_ = os.WriteFile(filepath.Join(dir, "Untitled.cuppa"), []byte("old"), 0o644)
	_, _ = ed.Add("lipgloss.box", 1, 1)
	f.SaveAs()
	press(f, true, false)
	clickButton(t, f, "Cancel")
	if b, _ := os.ReadFile(filepath.Join(dir, "Untitled.cuppa")); string(b) != "old" {
		t.Fatal("cancel must not overwrite")
	}
	f.SaveAs()
	press(f, true, false)
	clickButton(t, f, "Replace")
	if _, err := disk.Load(filepath.Join(dir, "Untitled.cuppa")); err != nil {
		t.Fatalf("replace failed: %v", err)
	}
}

func TestExportPlainText(t *testing.T) {
	f, ed, dir := setup(t)
	_, _ = ed.Add("lipgloss.box", 1, 1)
	f.ExportText(false)
	press(f, true, false)
	b, err := os.ReadFile(filepath.Join(dir, "Untitled.txt"))
	if err != nil || strings.Contains(string(b), "\x1b") || len(b) == 0 {
		t.Fatalf("export: %v %q", err, b)
	}
	if ed.Dirty() != true {
		t.Fatal("exporting must not mark the design saved")
	}
}

func withoutFreeze(f *Flow) {
	f.locate = func() (string, error) { return "", image.ErrFreezeMissing }
	f.CheckFreeze()
}

func TestExportImageWithoutFreezeExplainsInsteadOfAsking(t *testing.T) {
	f, _, dir := setup(t)
	withoutFreeze(f)
	if f.FreezeAvailable() {
		t.Fatal("freeze should be reported missing")
	}
	f.ExportImage(image.PNG)
	m := f.Modal()
	if m == nil {
		t.Fatal("expected the instructions popup")
	}
	body := bodyText(m.Lines())
	for _, want := range []string{"go install github.com/charmbracelet/freeze", "brew install", "CUPPA_FREEZE"} {
		if !strings.Contains(body, want) {
			t.Errorf("popup lacks %q:\n%s", want, body)
		}
	}
	press(f, true, false)
	if f.Modal() != nil || f.TakeJob() != nil {
		t.Fatal("no file dialog or job without freeze")
	}
	if _, err := os.Stat(filepath.Join(dir, "Untitled.png")); err == nil {
		t.Fatal("no picture should exist")
	}
}

func TestExportImageFindsFreezeInstalledMidSession(t *testing.T) {
	f, _, _ := setup(t)
	withoutFreeze(f)
	f.locate = func() (string, error) { return "freeze", nil }
	f.ExportImage(image.PNG)
	if f.Modal() == nil || !f.FreezeAvailable() {
		t.Fatal("a fresh check should find freeze and ask for a file name")
	}
}

func TestFreezeFailureAfterTheFactStillShowsAnError(t *testing.T) {
	f, _, _ := setup(t)
	f.Finish(os.ErrPermission)
	if f.Modal() == nil || f.Status() != "Export failed" {
		t.Fatal("expected a failure notice")
	}
}

func TestWelcomeShowsOnceAndOnlyWithoutFreeze(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "cfg", "freeze-notice")
	f, _, _ := setup(t)
	withoutFreeze(f)
	f.Welcome(marker)
	if f.Modal() == nil {
		t.Fatal("first run without freeze should show the notice")
	}
	press(f, true, false)
	f.Welcome(marker)
	if f.Modal() != nil {
		t.Fatal("the notice must only show once")
	}

	g, _, _ := setup(t)
	g.locate = func() (string, error) { return "freeze", nil }
	g.CheckFreeze()
	g.Welcome(filepath.Join(t.TempDir(), "other"))
	if g.Modal() != nil {
		t.Fatal("no notice when freeze is installed")
	}
}

func TestExportImageRealFreeze(t *testing.T) {
	if _, err := image.Locate(); err != nil {
		t.Skip("freeze not installed")
	}
	f, ed, dir := setup(t)
	_, _ = ed.Add("huh.confirm", 1, 1)
	f.ExportImage(image.SVG)
	press(f, true, false)
	job := f.TakeJob()
	if job == nil {
		t.Fatal("no job")
	}
	f.Finish(job.Run())
	if fi, err := os.Stat(filepath.Join(dir, "Untitled.svg")); err != nil || fi.Size() == 0 {
		t.Fatalf("no picture: %v", err)
	}
	if f.Status() != "Exported Untitled.svg" {
		t.Fatalf("status = %q", f.Status())
	}
}

func TestTitleFollowsTheFile(t *testing.T) {
	f, _, _ := setup(t)
	if f.Title() != "Untitled" {
		t.Fatal(f.Title())
	}
	f.path = filepath.Join("a", "demo.cuppa")
	if f.Title() != "demo.cuppa" {
		t.Fatal(f.Title())
	}
}

// bodyText joins a dialog's wrapped lines back into one sentence.
func bodyText(lines []string) string {
	var parts []string
	for _, l := range lines {
		l = strings.Trim(stripped(l), "│╭╰╮╯─ ")
		if l != "" {
			parts = append(parts, l)
		}
	}
	return strings.Join(parts, " ")
}

func TestOpenPathLoadsOrExplains(t *testing.T) {
	f, ed, dir := setup(t)
	doc := design.NewDocument("CLI", 33, 9)
	path, _ := disk.Save(filepath.Join(dir, "cli"), doc)
	if err := f.OpenPath(path); err != nil || ed.Document().Width != 33 || ed.Dirty() {
		t.Fatalf("open: %v", err)
	}
	bad := filepath.Join(dir, "bad.cuppa")
	_ = os.WriteFile(bad, []byte("zzz"), 0o644)
	if err := f.OpenPath(bad); err == nil || !strings.Contains(err.Error(), "not a Cuppa design") {
		t.Fatalf("bad: %v", err)
	}
}

func TestExportGoSourceWritesAProjectAndSaysWhatIsMissing(t *testing.T) {
	f, ed, dir := setup(t)
	if _, err := ed.Add("bubbles.textinput", 2, 2); err != nil {
		t.Fatal(err)
	}
	// A component the generator does not know (every catalog component has code).
	doc := ed.Document()
	doc.Add(design.Node{Component: "mystery.widget", Name: "Mystery", Rect: design.Rect{X: 2, Y: 6, W: 10, H: 3}})
	ed.Load(doc)
	f.ExportGoSource()
	if f.Modal() == nil {
		t.Fatal("a folder prompt should open")
	}
	want := filepath.Join(dir, "untitled-app")
	if !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "untitled-app") {
		t.Fatal("the folder is suggested next to the design")
	}
	press(f, true, false) // accept the suggestion
	if f.Modal() == nil || !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "Slices") {
		t.Fatal("the layout of the project is asked")
	}
	press(f, true, false) // One folder, the first button
	for _, name := range []string{"go.mod", "main.go", "layout.go", "runtime.go"} {
		if _, err := os.Stat(filepath.Join(want, name)); err != nil {
			t.Fatalf("%s not written: %v", name, err)
		}
	}
	if f.Modal() == nil || !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "mystery.widget") {
		t.Fatal("the notice lists what was not generated")
	}
	press(f, true, false)

	// A second export into the same folder is refused and nothing is overwritten.
	f.ExportGoSource()
	press(f, true, false)
	press(f, true, false) // One folder
	if f.Modal() == nil || !strings.Contains(stripped(strings.Join(f.Modal().Lines(), "\n")), "empty folder") {
		t.Fatal("exporting over an existing project is refused")
	}
}
