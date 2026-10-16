package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// categories says what each shipped component is for. The palette groups by
// this, not by who provides the component.
var categories = map[string]definition.Category{
	"lipgloss.rows": definition.CategoryLayout, "lipgloss.joinh": definition.CategoryLayout,
	"lipgloss.joinv": definition.CategoryLayout, "lipgloss.place": definition.CategoryLayout,
	"lipgloss.slot": definition.CategoryLayout, "community.flexbox": definition.CategoryLayout,
	"community.boxer": definition.CategoryLayout, "cuppa.divider": definition.CategoryLayout,
	"huh.form": definition.CategoryLayout,

	"lipgloss.box": definition.CategoryBox, "community.frame": definition.CategoryBox,
	"community.dialog": definition.CategoryBox, "community.overlay": definition.CategoryBox,
	"cuppa.panels": definition.CategoryBox, "cuppa.tabs": definition.CategoryBox,

	"lipgloss.label": definition.CategoryText, "community.bigtext": definition.CategoryText,
	"glamour.markdown": definition.CategoryText, "huh.note": definition.CategoryText,
	"bubbles.viewport": definition.CategoryText,

	"bubbles.textinput": definition.CategoryInput, "bubbles.textarea": definition.CategoryInput,
	"huh.input": definition.CategoryInput, "huh.text": definition.CategoryInput,
	"community.promptinput": definition.CategoryInput, "community.datepicker": definition.CategoryInput,
	"huh.confirm": definition.CategoryInput,

	"huh.select": definition.CategorySelect, "huh.multiselect": definition.CategorySelect,
	"community.dropdown": definition.CategorySelect, "community.promptselect": definition.CategorySelect,

	"bubbles.list": definition.CategoryData, "lipgloss.list": definition.CategoryData,
	"bubbles.table": definition.CategoryData, "lipgloss.table": definition.CategoryData,
	"community.bubbletable": definition.CategoryData, "bubbles.tree": definition.CategoryData,
	"lipgloss.tree": definition.CategoryData, "community.datatree": definition.CategoryData,

	"ntcharts.barchart": definition.CategoryCharts, "ntcharts.linechart": definition.CategoryCharts,
	"ntcharts.sparkline": definition.CategoryCharts, "ntcharts.streamline": definition.CategoryCharts,
	"ntcharts.timeseries": definition.CategoryCharts, "ntcharts.heatmap": definition.CategoryCharts,
	"ntcharts.canvas": definition.CategoryCharts, "ntcharts.chart3d": definition.CategoryCharts,

	"bubbles.spinner": definition.CategoryFeedback, "huh.spinner": definition.CategoryFeedback,
	"bubbles.progress": definition.CategoryFeedback, "lipgloss.progressline": definition.CategoryFeedback,
	"bubbles.timer": definition.CategoryFeedback, "bubbles.stopwatch": definition.CategoryFeedback,
	"community.toast": definition.CategoryFeedback, "community.statusmessage": definition.CategoryFeedback,

	"cuppa.menubar": definition.CategoryNavigation, "lipgloss.tabs": definition.CategoryNavigation,
	"lipgloss.keybar": definition.CategoryNavigation, "bubbles.help": definition.CategoryNavigation,
	"community.statusbar": definition.CategoryNavigation, "bubbles.paginator": definition.CategoryNavigation,

	"bubbles.filepicker": definition.CategoryFiles, "huh.filepicker": definition.CategoryFiles,
	"community.filetree": definition.CategoryFiles,

	"lipgloss.colourpicker": definition.CategoryColour, "lipgloss.swatch": definition.CategoryColour,

	"community.image": definition.CategoryMedia, "community.qrcode": definition.CategoryMedia,
	"community.pdfview": definition.CategoryMedia,
}

// withCategories gives each component its category (Other when unlisted).
func withCategories(defs []definition.Definition) []definition.Definition {
	for i := range defs {
		c, ok := categories[defs[i].ID]
		if !ok {
			c = definition.CategoryOther
		}
		defs[i].Category = c
	}
	return defs
}
