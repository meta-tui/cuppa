package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// ports says, for the components that show or take data, which property
// carries which kind of data. Two components can take each other's place in a
// design when the new one has the ports the old one's bindings use (see
// libs/catalog/compat): a text input and a colour picker both carry a value, a
// tab bar and a row of dialog buttons both carry items and the chosen one.
var ports = map[string]map[string]definition.Port{
	// A string.
	"lipgloss.label":          {"text": definition.PortValue},
	"lipgloss.place":          {"text": definition.PortValue},
	"lipgloss.swatch":         {"color": definition.PortValue},
	"lipgloss.colourpicker":   {"value": definition.PortValue},
	"bubbles.textinput":       {"value": definition.PortValue},
	"bubbles.textarea":        {"value": definition.PortValue},
	"huh.input":               {"value": definition.PortValue},
	"huh.text":                {"value": definition.PortValue},
	"community.promptinput":   {"value": definition.PortValue},
	"community.bigtext":       {"text": definition.PortValue},
	"community.qrcode":        {"content": definition.PortValue},
	"community.statusmessage": {"text": definition.PortValue},
	"community.toast":         {"message": definition.PortValue},
	"community.dialog":        {"text": definition.PortValue, "buttons": definition.PortItems, "active": definition.PortSelected},

	// Items to choose from, and the chosen one.
	"lipgloss.list":          {"items": definition.PortItems},
	"lipgloss.tabs":          {"tabs": definition.PortItems, "active": definition.PortSelected},
	"lipgloss.tree":          {"items": definition.PortItems},
	"bubbles.list":           {"items": definition.PortItems, "selected": definition.PortSelected},
	"bubbles.filepicker":     {"entries": definition.PortItems, "selected": definition.PortSelected},
	"huh.select":             {"options": definition.PortItems, "selected": definition.PortSelected},
	"huh.multiselect":        {"options": definition.PortItems},
	"huh.filepicker":         {"entries": definition.PortItems, "selected": definition.PortSelected},
	"community.dropdown":     {"options": definition.PortItems, "selected": definition.PortSelected},
	"community.promptselect": {"choices": definition.PortItems, "selected": definition.PortSelected},
	"community.filetree":     {"entries": definition.PortItems, "selected": definition.PortSelected},

	// Tables.
	"lipgloss.table":        {"headers": definition.PortHeaders, "rows": definition.PortRows},
	"lipgloss.rows":         {"rows": definition.PortRows},
	"bubbles.table":         {"columns": definition.PortHeaders, "rows": definition.PortRows, "selected": definition.PortSelected},
	"community.bubbletable": {"columns": definition.PortHeaders, "rows": definition.PortRows},

	// Text with a depth.
	"bubbles.tree":       {"items": definition.PortOutline, "selected": definition.PortSelected},
	"bubbles.viewport":   {"content": definition.PortOutline, "percent": definition.PortPercent},
	"community.datatree": {"data": definition.PortOutline},

	// Yes or no, and a percentage.
	"huh.confirm":      {"value": definition.PortChecked},
	"bubbles.progress": {"percent": definition.PortPercent},

	// Numbers.
	"ntcharts.barchart":   {"values": definition.PortNumbers, "labels": definition.PortLabels},
	"ntcharts.linechart":  {"values": definition.PortNumbers},
	"ntcharts.sparkline":  {"values": definition.PortNumbers},
	"ntcharts.streamline": {"values": definition.PortNumbers},
	"ntcharts.timeseries": {"values": definition.PortNumbers},
}

// withPorts sets the port of each property named in the table.
func withPorts(defs []definition.Definition) []definition.Definition {
	for i := range defs {
		byKey, ok := ports[defs[i].ID]
		if !ok {
			continue
		}
		props := make([]definition.PropSpec, len(defs[i].Props))
		copy(props, defs[i].Props)
		for j := range props {
			if port, ok := byKey[props[j].Key]; ok {
				props[j].Port = port
			}
		}
		defs[i].Props = props
	}
	return defs
}
