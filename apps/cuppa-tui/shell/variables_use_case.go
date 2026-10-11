package shell

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/confirm"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/variablesdialog"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// openVariables shows every named colour, input and event of the design.
func (m *Model) openVariables() {
	m.flow.Show(variablesdialog.New(m.ed.Variables()), m.afterVariables)
}

func (m *Model) afterVariables(o modal.Outcome) {
	switch o.Button {
	case variablesdialog.GoToButton:
		if o.Value == "" {
			// A screen key lives in the canvas options, which show when
			// nothing is selected.
			m.ed.Select()
			m.flow.SetStatus("Screen keys are in the details bar when nothing is selected")
			return
		}
		m.ed.Select(design.NodeID(o.Value))
	case variablesdialog.RenameButton:
		kind, name, _ := strings.Cut(o.Value, ":")
		m.flow.Show(prompt.New("Rename variable", "New name for "+name, name), func(p modal.Outcome) {
			if p.Canceled {
				return
			}
			if err := m.ed.RenameVariable(editor.VariableKind(kind), name, p.Value); err != nil {
				m.flow.Notice("Cannot rename", strings.TrimPrefix(err.Error(), "editor: "))
			}
		})
	}
}

// askQuestion shows a question with buttons and calls then with the one
// pressed; Esc and the Cancel button do nothing.
func (m *Model) askQuestion(title, body string, buttons []string, then func(string)) {
	m.flow.Show(confirm.New(title, body, buttons...), func(o modal.Outcome) {
		if o.Canceled || o.Button == "" || o.Button == "Cancel" {
			return
		}
		then(o.Button)
	})
}
