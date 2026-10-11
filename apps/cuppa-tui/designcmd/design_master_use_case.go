package designcmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// liftMaster moves the components that are the same on at least min pages to a
// master (ADR 0015). Two components are the same when their component, name,
// layout, properties and bindings are, apart from a property that is bound to
// an input (its value is the sample of that page). The pages that did not have
// one hide it. It reports what it lifted and what it left.
func liftMaster(pages []design.Page, min int, base *design.Document) ([]design.Node, []string) {
	type seen struct {
		key   string
		pages []int
		node  design.Node
	}
	var order []string
	groups := map[string]*seen{}
	for i, p := range pages {
		counted := map[string]bool{}
		for _, n := range p.Nodes {
			if n.IsContainer() || n.In != "" || len(n.Children) > 0 {
				continue
			}
			key := liftKey(n)
			if counted[key] {
				continue
			}
			counted[key] = true
			g, ok := groups[key]
			if !ok {
				g = &seen{key: key, node: n}
				groups[key] = g
				order = append(order, key)
			}
			g.pages = append(g.pages, i)
		}
	}
	var master []design.Node
	var report []string
	liftedAs := map[string]design.NodeID{}
	liftedNode := map[string]design.Node{}
	usedNames := map[string]design.Node{}
	for _, key := range order {
		g := groups[key]
		if len(g.pages) < min {
			continue
		}
		if lifted, taken := usedNames[g.node.Name]; taken {
			// A master keeps one component of a name, so the names in a layout
			// expression mean one component.
			report = append(report, fmt.Sprintf("kept on its %d pages: a second %q that differs from the lifted one in %s",
				len(g.pages), g.node.Name, strings.Join(differences(lifted, g.node), ", ")))
			continue
		}
		n := g.node.Clone()
		usedNames[g.node.Name] = g.node
		n.ID = base.NewID()
		master = append(master, n)
		liftedAs[key] = n.ID
		liftedNode[key] = n
		names := make([]string, len(g.pages))
		for k, pi := range g.pages {
			names[k] = pages[pi].Name
		}
		report = append(report, fmt.Sprintf("lifted %q (%s) from %d pages: %s", n.Name, n.Component, len(g.pages), strings.Join(names, ", ")))
	}
	// Take the lifted components off the pages that have them, and make the
	// others hide them.
	for i := range pages {
		have := map[string]bool{}
		var rest []design.Node
		for _, n := range pages[i].Nodes {
			key := liftKey(n)
			if master, ok := liftedNode[key]; ok && !n.IsContainer() && n.In == "" && len(n.Children) == 0 && !have[key] {
				have[key] = true
				// A bound property keeps this page's own text as the page's value.
				for prop, input := range n.Bind {
					if n.Props[prop] != master.Props[prop] {
						if pages[i].Values == nil {
							pages[i].Values = map[string]string{}
						}
						pages[i].Values[input] = n.Props[prop]
					}
				}
				continue
			}
			rest = append(rest, n)
		}
		pages[i].Nodes = rest
		for key, id := range liftedAs {
			if !have[key] {
				pages[i].Hide = append(pages[i].Hide, id)
			}
		}
		sort.Slice(pages[i].Hide, func(a, b int) bool { return pages[i].Hide[a] < pages[i].Hide[b] })
	}
	// What was left: components of one name that differ between pages, and
	// what differs.
	byName := map[string][]design.Node{}
	var names []string
	for _, p := range pages {
		for _, n := range p.Nodes {
			if _, ok := byName[n.Name]; !ok {
				names = append(names, n.Name)
			}
			byName[n.Name] = append(byName[n.Name], n)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		variants := byName[name]
		first := variants[0]
		differs := map[string]bool{}
		for _, other := range variants[1:] {
			for _, d := range differences(first, other) {
				differs[d] = true
			}
		}
		if len(differs) == 0 {
			continue
		}
		var list []string
		for d := range differs {
			list = append(list, d)
		}
		sort.Strings(list)
		report = append(report, fmt.Sprintf("left %q: differs in %s", name, strings.Join(list, ", ")))
	}
	if len(master) == 0 {
		report = append(report, fmt.Sprintf("nothing is on %d pages or more", min))
	}
	return master, report
}

// liftKey says when two components are the same: everything but the id, the
// rectangle of a component that has layout expressions, and the value of a
// bound property.
func liftKey(n design.Node) string {
	props := map[string]string{}
	for k, v := range n.Props {
		if _, bound := n.Bind[k]; !bound {
			props[k] = v
		}
	}
	rect := n.Rect
	if !n.Layout.IsZero() {
		rect = design.Rect{}
	}
	data, _ := json.Marshal(struct {
		Component, Name string
		Layout          design.Layout
		Rect            design.Rect
		Props, Bind     map[string]string
		ShowIf, Event   string
		Goto            string
		Hidden          bool
	}{n.Component, n.Name, n.Layout, rect, props, n.Bind, n.ShowIf, n.Event, n.Goto, n.Hidden})
	return string(data)
}

// differences names what differs between two components of one name.
func differences(a, b design.Node) []string {
	var out []string
	if a.Component != b.Component {
		out = append(out, "component")
	}
	if a.Layout != b.Layout {
		out = append(out, "layout")
	} else if a.Layout.IsZero() && a.Rect != b.Rect {
		out = append(out, "rect")
	}
	if a.ShowIf != b.ShowIf {
		out = append(out, "showIf")
	}
	if a.Event != b.Event {
		out = append(out, "event")
	}
	if a.Goto != b.Goto {
		out = append(out, "goto")
	}
	keys := map[string]bool{}
	for k := range a.Props {
		keys[k] = true
	}
	for k := range b.Props {
		keys[k] = true
	}
	for k := range keys {
		_, boundA := a.Bind[k]
		_, boundB := b.Bind[k]
		if boundA != boundB || !boundA && a.Props[k] != b.Props[k] || boundA && a.Bind[k] != b.Bind[k] {
			out = append(out, "prop "+k)
		}
	}
	return out
}
