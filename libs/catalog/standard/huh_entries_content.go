package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

const huhImport = "charm.land/huh/v2"

func huhEntries() []definition.Definition {
	title := func(def string) definition.PropSpec { return textProp("title", "Title", def) }
	desc := textProp("description", "Description", "")
	accent := colorProp("color", "Accent", "212")
	return []definition.Definition{
		{
			ID: "huh.input", Name: "Input", Family: definition.FamilyHuh,
			Description: "A single-line form field.",
			DefaultSize: size(40, 4), MinSize: size(16, 2), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("What is your name?"), desc, textProp("placeholder", "Placeholder", "Your name"), textProp("value", "Value", ""), accent},
		},
		{
			ID: "huh.text", Name: "Text", Family: definition.FamilyHuh,
			Description: "A multi-line form field.",
			DefaultSize: size(40, 7), MinSize: size(16, 3), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("Message"), desc, textProp("placeholder", "Placeholder", "Write something..."), textProp("value", "Value (| = new line)", ""), accent},
		},
		{
			ID: "huh.select", Name: "Select", Family: definition.FamilyHuh,
			Description: "Pick one option.",
			DefaultSize: size(32, 6), MinSize: size(14, 3), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("Pick a tea"), desc, textProp("options", "Options (comma separated)", "Earl Grey,Sencha,Rooibos"), intProp("selected", "Selected", 0, 0, 99), accent},
		},
		{
			ID: "huh.multiselect", Name: "Multi-select", Family: definition.FamilyHuh,
			Description: "Pick several options.",
			DefaultSize: size(32, 6), MinSize: size(14, 3), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("Extras"), desc, textProp("options", "Options (comma separated)", "Milk,Sugar,Lemon"), textProp("checked", "Checked indexes (comma separated)", "0"), intProp("cursor", "Cursor", 0, 0, 99), accent},
		},
		{
			ID: "huh.confirm", Name: "Confirm", Family: definition.FamilyHuh,
			Description: "A yes/no question.",
			DefaultSize: size(32, 4), MinSize: size(16, 2), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("Add milk?"), desc, textProp("affirmative", "Yes label", "Yes"), textProp("negative", "No label", "No"), boolProp("value", "Yes selected", true), accent},
		},
		{
			ID: "huh.note", Name: "Note", Family: definition.FamilyHuh,
			Description: "Read-only text inside a form.",
			DefaultSize: size(36, 5), MinSize: size(10, 2), Import: huhImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{title("Heads up"), textProp("body", "Body", "The kettle is hot."), accent},
		},
		{
			ID: "huh.spinner", Name: "Spinner", Family: definition.FamilyHuh,
			Description: "A titled spinner shown while an action runs.",
			DefaultSize: size(30, 1), MinSize: size(3, 1), Import: huhImport + "/spinner", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("title", "Title", "Steeping the tea..."),
				choiceProp("style", "Spinner", "dot", "line", "dot", "minidot", "jump", "pulse", "points", "globe", "moon", "monkey"),
				accent,
			},
		},
		{
			ID: "huh.filepicker", Name: "File picker", Family: definition.FamilyHuh,
			Description: "Choose a file inside a form.",
			DefaultSize: size(36, 10), MinSize: size(16, 4), Import: huhImport, Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{title("Choose a file"), textProp("entries", "Entries (comma separated, / = dir)", "docs/,libs/,README.md"), intProp("selected", "Selected", 0, 0, 99), accent},
		},
		{
			ID: "huh.form", Name: "Form group", Family: definition.FamilyHuh,
			Description: "A bordered group of fields with a title.",
			DefaultSize: size(48, 14), MinSize: size(20, 5), Import: huhImport, Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{title("Order"), textProp("footer", "Footer", "enter submit · tab next"), accent},
		},
	}
}
