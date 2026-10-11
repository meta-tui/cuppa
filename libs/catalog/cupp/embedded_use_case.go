package cupp

import (
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/catalog/registry"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Where components that only exist inside a design are listed.
const (
	EmbeddedPackID   = "from-design"
	EmbeddedPackName = "From this design"
)

// Embed returns doc with a copy of every pack component it uses (directly,
// inside groups, or inside other components), so the file opens correctly on a
// computer that does not have the packs.
func Embed(doc design.Document, cat Catalog) design.Document {
	seen := map[string]bool{}
	var out []design.Embedded
	var visit func(nodes []design.Node)
	visit = func(nodes []design.Node) {
		for _, n := range nodes {
			if n.IsGroup() {
				visit(n.Children)
				continue
			}
			def, ok := cat.Get(n.Component)
			if !ok || def.Inner == nil || seen[def.ID] {
				continue
			}
			seen[def.ID] = true
			out = append(out, design.Embedded{ID: def.ID, Composite: def.Inner.Clone()})
			visit(def.Inner.Nodes)
		}
	}
	visit(doc.Nodes)
	doc.Embedded = out
	return doc
}

// Adopt returns base plus the embedded components it does not already have,
// listed in the "From this design" pack. An installed pack always wins over a
// copy in a file.
func Adopt(base *registry.Registry, embedded []design.Embedded) *registry.Registry {
	var defs []definition.Definition
	for _, e := range embedded {
		if _, have := base.Get(e.ID); have || e.Composite.Check() != nil {
			continue
		}
		d := definitionOf(EmbeddedPackID, e.Composite)
		d.ID = e.ID
		defs = append(defs, d)
	}
	if len(defs) == 0 {
		return base
	}
	packs := append(base.Packs(), definition.Pack{ID: EmbeddedPackID, Name: EmbeddedPackName, Description: "Components saved inside the open design"})
	reg, err := registry.NewWithPacks(packs, append(base.List(), defs...)...)
	if err != nil {
		return base
	}
	return reg
}
