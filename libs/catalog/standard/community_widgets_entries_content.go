package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// communityWidgetEntries are components of community libraries found through
// the research on docs/catalog/community-components.md. Like the other
// community entries they are drawn as faithful approximations.
func communityWidgetEntries() []definition.Definition {
	accent := colorProp("color", "Color", "212")
	return []definition.Definition{
		{
			ID: "community.dropdown", Name: "Dropdown", Family: definition.FamilyCommunity,
			Description: "madicen/bubble-dropdown v2: a [ Label ▼ ] trigger that opens a panel of options.",
			DefaultSize: size(26, 7), MinSize: size(12, 1), Import: "github.com/madicen/bubble-dropdown/v2", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("options", "Options (comma separated)", "Earl Grey,Sencha,Rooibos,Matcha"),
				intProp("selected", "Selected option (-1 for none)", 1, -1, 99),
				textProp("placeholder", "Placeholder", "Pick a tea"),
				boolProp("open", "Show the open panel", true),
				accent,
			},
		},
		{
			ID: "community.promptinput", Name: "Prompt input", Family: definition.FamilyCommunity,
			Description: "erikgeiser/promptkit textinput: a prompt, an editable value, optional hidden mode and a validation error.",
			DefaultSize: size(36, 2), MinSize: size(8, 1), Import: "github.com/erikgeiser/promptkit/textinput", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("prompt", "Prompt", "Tea:"),
				textProp("value", "Value", ""),
				textProp("placeholder", "Placeholder", "Earl Grey"),
				boolProp("hidden", "Hidden (password)", false),
				textProp("error", "Validation error", ""),
				accent,
			},
		},
		{
			ID: "community.promptselect", Name: "Prompt select", Family: definition.FamilyCommunity,
			Description: "erikgeiser/promptkit selection: a prompt with a filter and a paginated list of choices.",
			DefaultSize: size(30, 6), MinSize: size(10, 2), Import: "github.com/erikgeiser/promptkit/selection", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("prompt", "Prompt", "Pick a tea:"),
				textProp("choices", "Choices (comma separated)", "Earl Grey,Sencha,Rooibos,Matcha,Chai,Oolong,Pu-erh"),
				intProp("selected", "Selected choice", 1, 0, 99),
				textProp("filter", "Filter", ""),
				accent,
			},
		},
		{
			ID: "community.datatree", Name: "Data tree", Family: definition.FamilyCommunity,
			Description: "Evertras/bubble-data-tree: any data structure as a tree of key: value lines (Bubble Tea v0.x era).",
			DefaultSize: size(34, 9), MinSize: size(10, 2), Import: "github.com/evertras/bubble-data-tree/datatree", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("data", "Data (| separates lines, two spaces nest)", "name: cuppa|version: 0.0.50|packs|  lipgloss: 9|  bubbles: 13|owners|  russo: admin"),
				accent,
			},
		},
		{
			ID: "community.pdfview", Name: "PDF viewer", Family: definition.FamilyCommunity,
			Description: "NimbleMarkets/ntcharts-pdf: a PDF page as text or as an image, with a status line.",
			DefaultSize: size(44, 14), MinSize: size(12, 3), Import: "github.com/NimbleMarkets/ntcharts-pdf/pdfview", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("path", "File", "tea-menu.pdf"),
				intProp("page", "Page", 3, 1, 999),
				intProp("pages", "Pages", 12, 1, 999),
				choiceProp("mode", "Mode", "text", "text", "image"),
				textProp("text", "Page text (| separates lines)", "Tea menu||Earl Grey      4.50|Sencha         5.00|Rooibos        4.00"),
			},
		},
		{
			ID: "ntcharts.chart3d", Name: "3D chart", Family: definition.FamilyNtcharts,
			Description: "NimbleMarkets/ntcharts3d: a scatter, surface, bar, line or vector-field series on three axes.",
			DefaultSize: size(44, 14), MinSize: size(16, 6), Import: "github.com/NimbleMarkets/ntcharts3d", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				choiceProp("kind", "Series", "surface", "scatter", "surface", "bar", "line", "vector"),
				textProp("title", "Title", "Steep temperature"),
				boolProp("legend", "Show legend", true),
				accent,
			},
		},
	}
}
