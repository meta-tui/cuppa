package component

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// ID is the catalog id of the component: its family and its name.
func (d Description) ID() string {
	family := d.Family
	if family == "" {
		family = "community"
	}
	return family + "." + d.Name
}

// Definition is the component as the catalog lists it. Its drawing in a design
// is the frame the preview asks for: the description carries no code, so a
// component that is not built into Cuppa is shown as a placeholder.
func (d Description) Definition() definition.Definition {
	family := definition.FamilyCommunity
	switch d.Family {
	case "lipgloss":
		family = definition.FamilyLipgloss
	case "bubbles":
		family = definition.FamilyBubbles
	case "huh":
		family = definition.FamilyHuh
	case "glamour":
		family = definition.FamilyGlamour
	case "ntcharts":
		family = definition.FamilyNtcharts
	}
	def := definition.Definition{
		ID: d.ID(), Name: d.Title, Family: family, Description: d.Description,
		DefaultSize: definition.Size{W: d.Size.Default.W, H: d.Size.Default.H},
		MinSize:     definition.Size{W: d.Size.Min.W, H: d.Size.Min.H},
		Import:      d.Go.Import, Status: definition.StatusPlaceholder,
	}
	kinds := map[string]definition.PropKind{
		"text": definition.PropText, "int": definition.PropInt, "float": definition.PropFloat, "bool": definition.PropBool,
		"color": definition.PropColor, "choice": definition.PropChoice,
	}
	for _, p := range d.Props {
		def.Props = append(def.Props, definition.PropSpec{
			Key: p.Key, Label: p.Label, Kind: kinds[p.Kind], Default: p.Default,
			Choices: p.Choices, Min: p.Min, Max: p.Max,
		})
	}
	return def
}
