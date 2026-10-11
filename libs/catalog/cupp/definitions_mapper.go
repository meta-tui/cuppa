package cupp

import (
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// ComponentID is the catalog id of a pack's component: "<pack>.<component>".
func ComponentID(pack, component string) string { return pack + "." + component }

// Definitions turns a pack into the catalog entry for the pack and one entry
// per component.
func Definitions(p Pack) (definition.Pack, []definition.Definition) {
	pack := definition.Pack{
		ID: definition.Family(p.ID), Name: p.Name, Version: p.Version, Description: p.Description, Source: p.Path, Builtin: p.Path == "",
	}
	defs := make([]definition.Definition, 0, len(p.Components))
	for _, c := range p.Components {
		defs = append(defs, definitionOf(pack.ID, c))
	}
	return pack, defs
}

func definitionOf(family definition.Family, c design.Composite) definition.Definition {
	props := make([]definition.PropSpec, len(c.Props))
	for i, e := range c.Props {
		props[i] = definition.PropSpec{
			Key: e.Key, Label: e.Label, Kind: definition.PropKind(e.Kind),
			Default: e.Default, Choices: e.Choices,
		}
	}
	inner := c.Clone()
	return definition.Definition{
		ID:          ComponentID(string(family), c.ID),
		Name:        c.Name,
		Family:      family,
		Description: c.Description,
		DefaultSize: definition.Size{W: c.W, H: c.H},
		MinSize:     definition.Size{W: min(c.W, 3), H: min(c.H, 1)},
		Props:       props,
		Status:      definition.StatusSupported,
		Inner:       &inner,
	}
}
