package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

const bubblesImport = "charm.land/bubbles/v2"

func bubblesEntries() []definition.Definition {
	return []definition.Definition{
		{
			ID: "bubbles.textinput", Name: "Text input", Family: definition.FamilyBubbles,
			Description: "A single-line text field.",
			DefaultSize: size(28, 1), MinSize: size(8, 1), Import: bubblesImport + "/textinput", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("prompt", "Prompt", "> "),
				textProp("placeholder", "Placeholder", "Type here"),
				textProp("value", "Value", ""),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.textarea", Name: "Text area", Family: definition.FamilyBubbles,
			Description: "A multi-line text editor.",
			DefaultSize: size(40, 6), MinSize: size(12, 3), Import: bubblesImport + "/textarea", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("placeholder", "Placeholder", "Write something..."),
				textProp("value", "Value (| = new line)", ""),
				boolProp("line_numbers", "Line numbers", true),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.list", Name: "List", Family: definition.FamilyBubbles,
			Description: "A browsable, filterable list with title, status bar and pagination.",
			DefaultSize: size(32, 10), MinSize: size(14, 5), Import: bubblesImport + "/list", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("title", "Title", "Pantry"),
				textProp("items", "Items (comma separated)", "Tea,Coffee,Biscuits,Milk,Sugar"),
				intProp("selected", "Selected index", 0, 0, 99),
				boolProp("show_filter", "Show filter", true),
				boolProp("show_status", "Show status bar", true),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.table", Name: "Table", Family: definition.FamilyBubbles,
			Description: "A table with a selectable row.",
			DefaultSize: size(40, 8), MinSize: size(14, 4), Import: bubblesImport + "/table", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("columns", "Columns (comma separated)", "Name,Kind,Qty"),
				textProp("rows", "Rows (; and , separated)", "Earl Grey,Black,2;Sencha,Green,1;Rooibos,Herbal,4"),
				intProp("selected", "Selected row", 0, 0, 99),
				colorProp("color", "Highlight color", "212"),
			},
		},
		{
			ID: "bubbles.tree", Name: "Tree", Family: definition.FamilyBubbles,
			Description: "A navigable tree with expandable branches (new in Bubbles v2.2).",
			DefaultSize: size(30, 10), MinSize: size(8, 2), Import: bubblesImport + "/tree", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("root", "Root", "cuppa"),
				textProp("items", "Items (| separated; two spaces indent)", "apps/|  cuppa-tui/|  cuppa-web/|libs/|  render/|go.work|README.md"),
				intProp("selected", "Selected row (0 = root)", 0, 0, 99),
				boolProp("show_help", "Show key help", true),
				colorProp("color", "Accent", "212"),
			},
		},
		{
			ID: "bubbles.viewport", Name: "Viewport", Family: definition.FamilyBubbles,
			Description: "A scrollable region of text.",
			DefaultSize: size(32, 8), MinSize: size(8, 3), Import: bubblesImport + "/viewport", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("content", "Content (| = new line)", "Scrollable content|second line|third line|fourth line|fifth line|sixth line|seventh line|eighth line|ninth line"),
				intProp("percent", "Scroll percent", 0, 0, 100),
				boolProp("show_scrollbar", "Scrollbar", true),
				colorProp("color", "Scrollbar color", "240"),
			},
		},
		{
			ID: "bubbles.paginator", Name: "Paginator", Family: definition.FamilyBubbles,
			Description: "Page indicator as dots or numbers.",
			DefaultSize: size(20, 1), MinSize: size(5, 1), Import: bubblesImport + "/paginator", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				intProp("total", "Pages", 5, 1, 20),
				intProp("page", "Current page", 1, 1, 20),
				choiceProp("style", "Style", "dots", "dots", "arabic"),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.filepicker", Name: "File picker", Family: definition.FamilyBubbles,
			Description: "A directory browser for choosing a file.",
			DefaultSize: size(34, 10), MinSize: size(14, 4), Import: bubblesImport + "/filepicker", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("path", "Directory", "~/projects/cuppa"),
				textProp("entries", "Entries (comma separated, / = dir)", "docs/,libs/,apps/,README.md,go.work"),
				intProp("selected", "Selected", 1, 0, 99),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.spinner", Name: "Spinner", Family: definition.FamilyBubbles,
			Description: "An animated activity indicator.",
			DefaultSize: size(16, 1), MinSize: size(2, 1), Import: bubblesImport + "/spinner", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				choiceProp("style", "Spinner", "dot", "line", "dot", "minidot", "jump", "pulse", "points", "globe", "moon", "monkey"),
				textProp("label", "Label", "Brewing..."),
				colorProp("color", "Color", "205"),
			},
		},
		{
			ID: "bubbles.progress", Name: "Progress", Family: definition.FamilyBubbles,
			Description: "A progress bar.",
			DefaultSize: size(30, 1), MinSize: size(6, 1), Import: bubblesImport + "/progress", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				intProp("percent", "Percent", 60, 0, 100),
				boolProp("show_percentage", "Show percentage", true),
				colorProp("color", "Fill color", "212"),
			},
		},
		{
			ID: "bubbles.timer", Name: "Timer", Family: definition.FamilyBubbles,
			Description: "A countdown timer display.",
			DefaultSize: size(24, 1), MinSize: size(5, 1), Import: bubblesImport + "/timer", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("value", "Remaining", "04:32"),
				textProp("label", "Label", "Steeping"),
				boolProp("running", "Running", true),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.stopwatch", Name: "Stopwatch", Family: definition.FamilyBubbles,
			Description: "An elapsed-time display.",
			DefaultSize: size(16, 1), MinSize: size(5, 1), Import: bubblesImport + "/stopwatch", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("value", "Elapsed", "00:12.3"),
				boolProp("running", "Running", true),
				colorProp("color", "Color", "212"),
			},
		},
		{
			ID: "bubbles.help", Name: "Help", Family: definition.FamilyBubbles,
			Description: "Key binding hints, short (one line) or expanded.",
			DefaultSize: size(50, 1), MinSize: size(10, 1), Import: bubblesImport + "/help", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("bindings", "Bindings (key:action, comma separated)", "↑/↓:move,enter:select,/:filter,q:quit"),
				boolProp("expanded", "Expanded", false),
				colorProp("color", "Key color", "212"),
			},
		},
	}
}
