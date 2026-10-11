package fileflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/meta-tui/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-tui/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-tui/cuppa/libs/cuppafile/format"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/export/gosource"
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
		project := gosource.GenerateWith(doc, f.cat, gosource.Options{Source: f.source(doc)})
		if err := gosource.Write(o.Value, project); err != nil {
			f.Notice("Cannot export Go source", err.Error())
			return
		}
		f.status = "Exported Go source to " + o.Value
		body := fmt.Sprintf("Written to %s\n\nNext: go mod tidy && go run .", o.Value)
		if len(project.Notes) > 0 {
			body += "\n\nNot generated yet:\n- " + strings.Join(project.Notes, "\n- ")
		}
		f.Notice("Go source exported", body)
	})
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
