package fileflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/confirm"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
)

// ExportGoSource writes the design as Go source where the design says it goes.
// The first time (or on "Change…") it asks for the folder and the layout and
// records them in the design, so the design file is the one place that knows
// the structure of the program: exporting again needs no questions.
func (f *Flow) ExportGoSource() {
	s := f.ed.ExportSettings()
	if !s.IsSet() {
		f.chooseExport(s)
		return
	}
	body := fmt.Sprintf("%s\n\nFolder: %s", describeLayout(s), gosource.ExportDir(s, f.path))
	if s.Module != "" {
		body += "\nModule: " + s.Module
	}
	f.show(confirm.New("Export Go source", body, "Export", "Change…", "Cancel"), func(o modal.Outcome) {
		switch o.Button {
		case "Export":
			f.runExport(s)
		case "Change…":
			f.chooseExport(s)
		}
	})
}

func describeLayout(s design.ExportSettings) string {
	switch {
	case s.Library:
		return "Library: the pages in their folders, the runtime in internal/cuppa, no go.mod or main.go"
	case s.Layout == design.ExportSlices:
		return "Slices: a folder and a package per page, the router in app/, main.go at the root"
	}
	return "One folder: the screens, the router and the runtime side by side"
}

// chooseExport asks for the folder, the layout and the module, records them in
// the design and exports.
func (f *Flow) chooseExport(prev design.ExportSettings) {
	doc := f.ed.Document()
	suggestion := ""
	if prev.Dir != "" {
		suggestion = gosource.ExportDir(prev, f.path)
	} else {
		start := f.startDir()
		if start == "" {
			start = "."
		}
		suggestion = filepath.Join(start, gosource.Slug(doc.Name)+"-app")
	}
	f.show(prompt.New("Export Go source", "Folder for the files", suggestion), func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		dir := strings.TrimSpace(o.Value)
		body := "One folder: the screens, the router and the runtime side by side.\n\n" +
			"Slices: a folder and a package per page, the router in app/, the runtime in internal/cuppa and main.go at the root.\n\n" +
			"Library: the pages in their folders (see Folder… in the pages list) inside a module you already have; no go.mod, main.go or router."
		f.show(confirm.New("Layout of the project", body, "One folder", "Slices", "Library", "Cancel"), func(o modal.Outcome) {
			switch o.Button {
			case "One folder":
				f.record(design.ExportSettings{Layout: design.ExportFlat, Dir: f.storedDir(dir)})
			case "Slices", "Library":
				library := o.Button == "Library"
				label := "Module path (go.mod)"
				initial := prev.Module
				if library {
					label = "Import path of this folder"
				}
				if initial == "" && !library {
					initial = gosource.Slug(doc.Name)
				}
				f.show(prompt.New("Export Go source", label, initial), func(o modal.Outcome) {
					if o.Canceled {
						return
					}
					f.record(design.ExportSettings{Layout: design.ExportSlices, Library: library, Dir: f.storedDir(dir), Module: strings.TrimSpace(o.Value)})
				})
			}
		})
	})
}

// record stores the settings in the design and exports with them.
func (f *Flow) record(s design.ExportSettings) {
	if err := f.ed.SetExportSettings(s); err != nil {
		f.Notice("Cannot export Go source", strings.TrimPrefix(err.Error(), "editor: "))
		return
	}
	f.runExport(f.ed.ExportSettings())
}

// storedDir is the folder as the design keeps it: relative to the design file
// when it has one, so the design and the program can move together.
func (f *Flow) storedDir(dir string) string {
	if f.path == "" {
		return dir
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	absFile, err := filepath.Abs(f.path)
	if err != nil {
		return dir
	}
	rel, err := filepath.Rel(filepath.Dir(absFile), absDir)
	if err != nil {
		return dir
	}
	return filepath.ToSlash(rel)
}

// runExport generates the project the settings ask for and writes it.
func (f *Flow) runExport(s design.ExportSettings) {
	doc := f.ed.Document()
	dir := gosource.ExportDir(s, f.path)
	project, err := gosource.GenerateExport(doc, f.cat, s, f.source(doc))
	if err != nil {
		f.Notice("Cannot export Go source", err.Error())
		return
	}
	if err := gosource.WriteExport(dir, s, project); err != nil {
		f.Notice("Cannot export Go source", err.Error())
		return
	}
	f.status = "Exported Go source to " + dir
	body := fmt.Sprintf("Written to %s\n\n", dir)
	if s.Library {
		body += "Your module imports the pages from " + s.Module + "."
	} else {
		body += "Next: go mod tidy && go run ."
	}
	if len(project.Requires) > 0 && s.Library {
		body += "\n\nNeeds in go.mod:\n- " + strings.Join(project.Requires, "\n- ")
	}
	if len(project.Notes) > 0 {
		body += "\n\nNot generated yet:\n- " + strings.Join(project.Notes, "\n- ")
	}
	f.Notice("Go source exported", body)
}

// CheckGoExport tells whether the Go files in the folder the design names are
// what the design exports: nothing is written.
func (f *Flow) CheckGoExport() {
	s := f.ed.ExportSettings()
	if !s.IsSet() {
		f.Notice("Check Go export", "This design has not been exported yet. Use Export, Go source first: it records the folder in the design.")
		return
	}
	doc := f.ed.Document()
	dir := gosource.ExportDir(s, f.path)
	project, err := gosource.GenerateExport(doc, f.cat, s, f.source(doc))
	if err != nil {
		f.Notice("Check Go export", err.Error())
		return
	}
	diffs := gosource.CheckProject(dir, project)
	if len(diffs) == 0 {
		f.status = "Go export is current"
		f.Notice("Check Go export", fmt.Sprintf("%s is current: %d files.", dir, len(project.Files)))
		return
	}
	shown := diffs
	more := ""
	if len(shown) > 12 {
		more = fmt.Sprintf("\n… and %d more", len(shown)-12)
		shown = shown[:12]
	}
	f.status = fmt.Sprintf("Go export: %d files differ", len(diffs))
	f.Notice("Check Go export", fmt.Sprintf("%s is out of date: %d files differ.\n\n%s%s\n\nExport again to update it.", dir, len(diffs), strings.Join(shown, "\n"), more))
}

// source is the design as the bytes of a .cuppa file, for an export to carry
// so the application can be opened again from its Go.
func (f *Flow) source(doc design.Document) []byte {
	data, err := format.Encode(doc)
	if err != nil {
		return nil
	}
	return data
}
