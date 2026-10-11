// Package registry stores component definitions and answers lookups.
package registry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

// Registry is an immutable set of component definitions and the packs they
// belong to.
type Registry struct {
	byID  map[string]definition.Definition
	list  []definition.Definition
	packs []definition.Pack
}

// New builds a registry of the built-in families. It fails on an empty or
// duplicate id.
func New(defs ...definition.Definition) (*Registry, error) {
	return NewWithPacks(nil, defs...)
}

// NewWithPacks builds a registry whose components belong to packs. A family
// with no pack gets a built-in one named after it, so components never
// disappear. Packs keep the order given, built-in families first.
func NewWithPacks(packs []definition.Pack, defs ...definition.Definition) (*Registry, error) {
	r := &Registry{byID: make(map[string]definition.Definition, len(defs))}
	seen := map[definition.Family]bool{}
	for _, p := range packs {
		if p.ID == "" {
			return nil, fmt.Errorf("registry: pack %q has no id", p.Name)
		}
		if seen[p.ID] {
			return nil, fmt.Errorf("registry: duplicate pack %q", p.ID)
		}
		seen[p.ID] = true
		r.packs = append(r.packs, p)
	}
	used := map[definition.Family]bool{}
	for _, d := range defs {
		if !d.Hidden {
			used[d.Family] = true
		}
	}
	for _, f := range definition.FamilyOrder { // built-in families without a pack
		if used[f] && !seen[f] {
			seen[f] = true
			r.packs = append(r.packs, definition.Pack{ID: f, Name: f.Title(), Builtin: true})
		}
	}
	for _, d := range defs {
		if d.Family != "" && !d.Hidden && !seen[d.Family] {
			seen[d.Family] = true
			r.packs = append(r.packs, definition.Pack{ID: d.Family, Name: d.Family.Title(), Builtin: true})
		}
	}
	for _, d := range defs {
		if d.ID == "" {
			return nil, fmt.Errorf("registry: definition %q has no id", d.Name)
		}
		if _, dup := r.byID[d.ID]; dup {
			return nil, fmt.Errorf("registry: duplicate id %q", d.ID)
		}
		r.byID[d.ID] = d
		r.list = append(r.list, d)
	}
	order := make(map[definition.Family]int, len(r.packs))
	for i, p := range r.packs {
		order[p.ID] = i
	}
	sort.SliceStable(r.list, func(i, j int) bool {
		a, b := r.list[i], r.list[j]
		if order[a.Family] != order[b.Family] {
			return order[a.Family] < order[b.Family]
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	return r, nil
}

// Get returns the definition with the given id.
func (r *Registry) Get(id string) (definition.Definition, bool) {
	d, ok := r.byID[id]
	return d, ok
}

// List returns every definition ordered by family, then name.
func (r *Registry) List() []definition.Definition {
	return append([]definition.Definition(nil), r.list...)
}

// Families returns the packs that have at least one definition, in display order.
func (r *Registry) Families() []definition.Family {
	var out []definition.Family
	for _, p := range r.packs {
		if len(r.ByFamily(p.ID)) > 0 {
			out = append(out, p.ID)
		}
	}
	return out
}

// Packs returns every pack, in display order, including empty ones.
func (r *Registry) Packs() []definition.Pack {
	return append([]definition.Pack(nil), r.packs...)
}

// Title is the display name of a pack.
func (r *Registry) Title(f definition.Family) string {
	for _, p := range r.packs {
		if p.ID == f {
			return p.Name
		}
	}
	return f.Title()
}

// Only returns a registry with just the packs keep accepts. The original is
// untouched, so designs can still resolve components of a pack that is off.
func (r *Registry) Only(keep func(definition.Family) bool) *Registry {
	out := &Registry{byID: map[string]definition.Definition{}}
	for _, p := range r.packs {
		if keep(p.ID) {
			out.packs = append(out.packs, p)
		}
	}
	for _, d := range r.list {
		if d.Hidden || keep(d.Family) {
			out.byID[d.ID] = d
			out.list = append(out.list, d)
		}
	}
	return out
}

// ByFamily returns the definitions of one family ordered by name.
func (r *Registry) ByFamily(f definition.Family) []definition.Definition {
	var out []definition.Definition
	for _, d := range r.list {
		if d.Family == f && !d.Hidden {
			out = append(out, d)
		}
	}
	return out
}

// Search returns the definitions whose name, id or description contain query,
// ignoring case. An empty query returns everything.
func (r *Registry) Search(query string) []definition.Definition {
	q := strings.ToLower(strings.TrimSpace(query))
	var out []definition.Definition
	if q == "" {
		for _, d := range r.list {
			if !d.Hidden {
				out = append(out, d)
			}
		}
		return out
	}
	for _, d := range r.list {
		if d.Hidden {
			continue
		}
		hay := strings.ToLower(d.Name + " " + d.ID + " " + d.Description)
		if strings.Contains(hay, q) {
			out = append(out, d)
		}
	}
	return out
}
