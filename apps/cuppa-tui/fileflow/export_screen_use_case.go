package fileflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/meta-tui/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-tui/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/export/gosource"
)

// ExportScreen asks for a folder and writes the design there as a screen of a
// Go package: its contract, its view and the runtime they share. Other screens
// already in the folder stay.
func (f *Flow) ExportScreen() {
	doc := f.ed.Document()
	start := f.startDir()
	if start == "" {
		start = "."
	}
	f.show(prompt.New("Export Go screen", "Folder of the screens package", filepath.Join(start, "screens")), func(o modal.Outcome) {
		if o.Canceled {
			return
		}
		project := gosource.GenerateScreensWith([]design.Document{doc}, f.cat, filepath.Base(o.Value), gosource.Options{Source: f.source(doc)})
		if err := gosource.WriteScreens(o.Value, project); err != nil {
			f.Notice("Cannot export Go screen", err.Error())
			return
		}
		f.status = "Exported Go screen to " + o.Value
		body := fmt.Sprintf("Written to %s\n\nThe package holds only generated files; call %s(props, w, h) from your program.", o.Value, gosource.ScreenName(doc.Name))
		if len(project.Requires) > 0 {
			body += "\n\nAdd to go.mod:\n- " + strings.Join(project.Requires, "\n- ")
		}
		if len(project.Notes) > 0 {
			body += "\n\nNot exported:\n- " + strings.Join(project.Notes, "\n- ")
		}
		f.Notice("Go screen exported", body)
	})
}
