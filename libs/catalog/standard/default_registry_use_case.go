package standard

import (
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/catalog/registry"
)

// All returns every component definition Cuppa ships with.
func All() []definition.Definition {
	var defs []definition.Definition
	defs = append(defs, lipglossEntries()...)
	defs = append(defs, drawEntries()...)
	defs = append(defs, bubblesEntries()...)
	defs = append(defs, huhEntries()...)
	defs = append(defs, glamourEntries()...)
	defs = append(defs, ntchartsEntries()...)
	defs = append(defs, communityEntries()...)
	defs = append(defs, communityWidgetEntries()...)
	return withCategories(withRoles(defs))
}

// Default builds the registry of shipped components. It panics if the shipped
// definitions are inconsistent, which is a programming error caught by tests.
func Default() *registry.Registry {
	r, err := registry.NewWithPacks(Packs(), All()...)
	if err != nil {
		panic(err)
	}
	return r
}

// colourRoles names, for the components whose "color" is not the accent, the
// part of the theme it follows. Every other component's "color" is the accent
// (the secondary colour).
var colourRoles = map[string]definition.Role{
	"lipgloss.box": definition.RoleBorder, "lipgloss.tabs": definition.RoleBorder, "lipgloss.table": definition.RoleBorder,
	"community.frame": definition.RoleBorder, "community.dialog": definition.RoleBorder, "community.overlay": definition.RoleBorder,
	"community.flexbox": definition.RoleBorder, "community.boxer": definition.RoleBorder, "huh.form": definition.RoleBorder,
	"cuppa.divider": definition.RoleBorder,
	"lipgloss.label": definition.RoleText, "lipgloss.list": definition.RoleText, "lipgloss.tree": definition.RoleText,
}

// withRoles ties each component's "color" property to its theme role.
func withRoles(defs []definition.Definition) []definition.Definition {
	for i := range defs {
		role, ok := colourRoles[defs[i].ID]
		if !ok {
			role = definition.RoleSecondary
		}
		props := make([]definition.PropSpec, len(defs[i].Props))
		copy(props, defs[i].Props)
		for j := range props {
			if props[j].Kind == definition.PropColor && props[j].Key == "color" {
				props[j].Role = role
			}
		}
		defs[i].Props = props
	}
	return defs
}
