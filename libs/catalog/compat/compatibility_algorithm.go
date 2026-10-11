// Package compat decides whether one component can take another's place in a
// design without losing the variables bound to it. Components declare ports on
// their properties (definition.Port); a change keeps the bindings and values of
// the properties whose ports the new component has too.
package compat

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// Move is a property of the old component whose value and binding go to a
// property of the new one.
type Move struct {
	From, To string
	Port     definition.Port
}

// Report is what changing one component for another does.
type Report struct {
	// Moves are the properties that carry over, by port.
	Moves []Move
	// Lost are the bound properties of the old component that have no
	// counterpart: their bindings would be removed.
	Lost []string
	// Added are the ports of the new component the old one did not have, as the
	// new component's property keys, free to be bound.
	Added []string
}

// Compatible is true when no binding would be lost.
func (r Report) Compatible() bool { return len(r.Lost) == 0 }

// Check says what changing from for to does when the properties named in bound
// are bound to inputs.
func Check(from, to definition.Definition, bound []string) Report {
	var r Report
	taken := map[string]bool{}
	for _, p := range from.Props {
		if p.Port == definition.PortNone {
			continue
		}
		target, ok := to.PropOfPort(p.Port)
		if !ok || taken[target.Key] {
			continue
		}
		taken[target.Key] = true
		r.Moves = append(r.Moves, Move{From: p.Key, To: target.Key, Port: p.Port})
	}
	moved := map[string]bool{}
	for _, m := range r.Moves {
		moved[m.From] = true
	}
	for _, key := range bound {
		if !moved[key] {
			r.Lost = append(r.Lost, key)
		}
	}
	for _, p := range to.Props {
		if p.Port != definition.PortNone && !taken[p.Key] {
			r.Added = append(r.Added, p.Key)
		}
	}
	return r
}

// Candidates are the components that can take the place of from without losing
// a binding, in the order of all. A component that has nothing bound can be
// changed for one that shares a port with it; one that has bindings, for one
// that carries every port they use. A pack component, a group and from itself
// are never candidates.
func Candidates(all []definition.Definition, from definition.Definition, bound []string) []definition.Definition {
	var out []definition.Definition
	for _, to := range all {
		if to.ID == from.ID || to.Inner != nil || len(to.Ports()) == 0 {
			continue
		}
		r := Check(from, to, bound)
		if !r.Compatible() {
			continue
		}
		if len(bound) == 0 && len(r.Moves) == 0 {
			continue
		}
		out = append(out, to)
	}
	return out
}
