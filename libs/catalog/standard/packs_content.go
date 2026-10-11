package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// Packs describes the built-in packs, in palette order.
func Packs() []definition.Pack {
	return []definition.Pack{
		{ID: definition.FamilyLipgloss, Name: "Lip Gloss", Description: "Boxes, labels, lists, tables, tabs and layout helpers", Builtin: true},
		{ID: definition.FamilyBubbles, Name: "Bubbles", Description: "The official Bubble Tea components", Builtin: true},
		{ID: definition.FamilyHuh, Name: "Huh forms", Description: "Form fields and groups from Huh", Builtin: true},
		{ID: definition.FamilyGlamour, Name: "Glamour", Description: "Rendered markdown", Builtin: true},
		{ID: definition.FamilyNtcharts, Name: "ntcharts", Description: "Terminal charts", Builtin: true},
		{ID: definition.FamilyCommunity, Name: "Community", Description: "Unofficial components from the Bubble Tea ecosystem", Builtin: true},
	}
}
