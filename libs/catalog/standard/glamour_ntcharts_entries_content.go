package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

const (
	glamourImport  = "charm.land/glamour/v2"
	ntchartsImport = "github.com/NimbleMarkets/ntcharts/v2"
)

func glamourEntries() []definition.Definition {
	return []definition.Definition{
		{
			ID: "glamour.markdown", Name: "Markdown", Family: definition.FamilyGlamour,
			Description: "Rendered markdown (headings, lists, code, bold).",
			DefaultSize: size(40, 10), MinSize: size(10, 2), Import: glamourImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("markdown", "Markdown (| = new line)", "# Tea time|Some **bold** text and `code`.|- Earl Grey|- Sencha|> steep for 3 minutes"),
				choiceProp("style", "Style", "dark", "dark", "light", "pink", "notty"),
			},
		},
	}
}

func ntchartsEntries() []definition.Definition {
	values := func(def string) definition.PropSpec { return textProp("values", "Values (comma separated)", def) }
	accent := colorProp("color", "Color", "212")
	return []definition.Definition{
		{
			ID: "ntcharts.barchart", Name: "Bar chart", Family: definition.FamilyNtcharts,
			Description: "Vertical bars with labels.",
			DefaultSize: size(36, 10), MinSize: size(10, 4), Import: ntchartsImport + "/barchart", Status: definition.StatusSupported,
			Props: []definition.PropSpec{values("3,7,4,9,5"), textProp("labels", "Labels (comma separated)", "Mon,Tue,Wed,Thu,Fri"), accent},
		},
		{
			ID: "ntcharts.linechart", Name: "Line chart", Family: definition.FamilyNtcharts,
			Description: "A line chart with axes.",
			DefaultSize: size(40, 10), MinSize: size(10, 4), Import: ntchartsImport + "/linechart", Status: definition.StatusSupported,
			Props: []definition.PropSpec{values("1,3,2,5,4,6,5,8"), boolProp("axes", "Show axes", true), accent},
		},
		{
			ID: "ntcharts.sparkline", Name: "Sparkline", Family: definition.FamilyNtcharts,
			Description: "A compact one-line trend.",
			DefaultSize: size(30, 1), MinSize: size(4, 1), Import: ntchartsImport + "/sparkline", Status: definition.StatusSupported,
			Props: []definition.PropSpec{values("2,4,3,6,8,5,7,9,6,4"), accent},
		},
		{
			ID: "ntcharts.streamline", Name: "Streamline", Family: definition.FamilyNtcharts,
			Description: "A streaming line chart.",
			DefaultSize: size(30, 4), MinSize: size(6, 2), Import: ntchartsImport + "/streamlinechart", Status: definition.StatusSupported,
			Props: []definition.PropSpec{values("1,2,4,3,5,7,6,8,9,7"), accent},
		},
		{
			ID: "ntcharts.timeseries", Name: "Time series", Family: definition.FamilyNtcharts,
			Description: "Values over time.",
			DefaultSize: size(40, 10), MinSize: size(10, 4), Import: ntchartsImport + "/timeserieslinechart", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{values("4,5,3,6,8,7,9,6,8"), boolProp("axes", "Show axes", true), accent},
		},
		{
			ID: "ntcharts.heatmap", Name: "Heat map", Family: definition.FamilyNtcharts,
			Description: "A grid shaded by value.",
			DefaultSize: size(30, 8), MinSize: size(6, 3), Import: ntchartsImport + "/heatmap", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{intProp("seed", "Pattern seed", 3, 0, 99), accent},
		},
		{
			ID: "ntcharts.canvas", Name: "Chart canvas", Family: definition.FamilyNtcharts,
			Description: "An empty plotting area with axes.",
			DefaultSize: size(30, 8), MinSize: size(6, 3), Import: ntchartsImport + "/canvas", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{textProp("title", "Title", "canvas"), accent},
		},
	}
}
