package fileflow

import (
	"fmt"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/confirm"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
)

// ExportGoSource asks for a folder and writes the design there as a Go
// program that uses the real components.
func (f *Flow) ExportGoSource() {
	doc := f.ed.Document()
	start := f.startDir()
	if start == "" {
		start = "."
	}
	suggestion := filepath.Join(start, gosource.Slug(doc.Name)+"-app")
	f.show(prompt.New("Export Go source", "Folder for the new project", suggestion), func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		dir := o.Value
		body := "One folder: the screens, the router and the runtime side by side.\n\n" +
			"Slices: a folder and a package per page (pages/<page>/), the router in app/, the runtime in internal/cuppa/ and main.go at the root."
		f.show(confirm.New("Layout of the project", body, "One folder", "Slices", "Cancel"), func(o modal.Outcome) {
			switch o.Button {
			case "One folder":
				f.writeGoSource(dir, doc, gosource.Options{Source: f.source(doc)}, false)
			case "Slices":
				f.show(prompt.New("Export Go source", "Module path (go.mod)", gosource.Slug(doc.Name)), func(o modal.Outcome) {
					if o.Canceled {
						return
					}
					f.writeGoSource(dir, doc, gosource.Options{Source: f.source(doc), Module: strings.TrimSpace(o.Value)}, true)
				})
			}
		})
	})
}

// writeGoSource generates the project in the chosen layout and writes it.
func (f *Flow) writeGoSource(dir string, doc design.Document, opts gosource.Options, slices bool) {
	var project gosource.Project
	if slices {
		var err error
		if project, err = gosource.GenerateSlices(doc, f.cat, opts); err != nil {
			f.Notice("Cannot export Go source", err.Error())
			return
		}
	} else {
		project = gosource.GenerateWith(doc, f.cat, opts)
	}
	write := gosource.Write
	if slices {
		// Exporting again replaces what Cuppa wrote and keeps the person's files.
		write = gosource.ReplaceProject
	}
	if err := write(dir, project); err != nil {
		f.Notice("Cannot export Go source", err.Error())
		return
	}
	f.status = "Exported Go source to " + dir
	body := fmt.Sprintf("Written to %s\n\nNext: go mod tidy && go run .", dir)
	if len(project.Notes) > 0 {
		body += "\n\nNot generated yet:\n- " + strings.Join(project.Notes, "\n- ")
	}
	f.Notice("Go source exported", body)
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
