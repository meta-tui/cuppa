package gosource

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// The master's parts written once (ADR 0015, step 2). The flat generator
// writes every page with the master's components as parts of its own. In the
// slice layout the master is a package: its parts are generated once, typed on
// the master's props, and a page takes them through a function that turns its
// own props into the master's. What a page draws is the same by construction:
// the parts are the text the page would have had, in the same order.

// masterIdent is the screen the master is generated as.
const (
	masterIdent = "Master"
	masterLower = "master"
	// masterPackage is the package the master's parts are in, and the name a
	// page imports it as.
	masterPackage = "master"
)

// masterShare is the master generated as a screen of its own.
type masterShare struct {
	sc screen
	// events are the events the master's parts raise, as the package names them.
	events map[string]bool
}

// masterScreen builds the master of a design as a screen, or nil when the
// design has none or the master cannot be a package of its own: a container or
// a colour picker keeps its state in code the page owns.
func masterScreen(doc design.Document, cat Catalog) (*masterShare, []string) {
	nodes := doc.MasterNodes()
	if len(nodes) == 0 {
		return nil, nil
	}
	d := doc.Clone()
	m := d.MasterIndex()
	if m != d.ActivePage() {
		d.Samples = d.Pages[m].Samples
	}
	d.Pages, d.Page = nil, 0
	d.Name, d.Folder = masterIdent, ""
	d.Keys = nil
	d.Nodes = design.Document{Nodes: nodes}.Clone().Nodes
	sc := buildScreen(d, masterIdent, leaves(d, cat), cat)
	if !sc.sharable {
		return nil, []string{"the master has a container or a colour picker: each page keeps its own copy of the master's parts"}
	}
	events := map[string]bool{}
	for _, l := range leaves(d, cat) {
		if l.Event != "" {
			events[pascal(l.Event)] = true
		}
	}
	return &masterShare{sc: sc.with("cuppa"), events: events}, nil
}

// masterSource is the source of the master's package in the flat naming, which
// the slice transform turns into the package: the screen and the function a
// page takes the parts through.
func (m *masterShare) source() string {
	body := m.sc.contract
	if at := strings.Index(body, "package cuppa"); at >= 0 {
		body = body[at+len("package cuppa"):]
	}
	return strings.TrimRight(m.sc.view, "\n") + "\n" + body + `
// MasterPartsFor is the master's parts for a page whose props are P: conv gives
// the master's props from the page's, and hide names the parts the page does
// not show.
func MasterPartsFor[P any](conv func(P) MasterProps, hide ...string) []screenPart[P] {
	out := make([]screenPart[P], 0, len(masterParts))
	for _, part := range masterParts {
		hidden := false
		for _, name := range hide {
			hidden = hidden || name == part.Name
		}
		if hidden {
			continue
		}
		shared := screenPart[P]{placed: part.placed, Roles: part.Roles, Event: part.Event}
		if late := part.Late; late != nil {
			shared.Late = func(p P, v map[string]string, w, h int) { late(conv(p), v, w, h) }
		}
		if apply := part.Apply; apply != nil {
			shared.Apply = func(p P, v map[string]string, w, h int) bool { return apply(conv(p), v, w, h) }
		}
		out = append(out, shared)
	}
	return out
}
`
}

// share rewrites a page's screen so that the master's parts it shows come from
// the master's package. It returns the page unchanged, and says why, when the
// page's parts are not the master's.
func (m *masterShare) share(sc screen, doc design.Document, ident string, cat Catalog) (screen, string) {
	var own []design.Node
	for _, n := range doc.Nodes {
		if strings.HasPrefix(string(n.ID), design.MasterPrefix) {
			own = append(own, n)
		}
	}
	if len(own) == 0 {
		return sc, ""
	}
	count := len(leaves(design.Document{Width: doc.Width, Height: doc.Height, Theme: doc.Theme, Background: doc.Background, Nodes: own}, cat))
	fail := func(why string) (screen, string) {
		return sc, fmt.Sprintf("%s keeps its own copy of the master's parts: %s", doc.Name, why)
	}
	if !sc.sharable || count == 0 || count > len(sc.partNames) {
		return fail("it has a container or a colour picker")
	}
	// The page's first parts are the master's, in the master's order.
	masterAt := map[string]int{}
	for i, n := range m.sc.partNames {
		masterAt[n] = i
	}
	last := -1
	shown := map[string]bool{}
	for _, n := range sc.partNames[:count] {
		i, ok := masterAt[n]
		if !ok || i <= last {
			return fail("its parts are not the master's")
		}
		last = i
		shown[n] = true
	}
	var hidden []string
	for _, n := range m.sc.partNames {
		if !shown[n] {
			hidden = append(hidden, strconv.Quote(n))
		}
	}
	// What the master's parts read must be the page's inputs of the same kind.
	pageInput := map[string]input{}
	for _, in := range sc.inputs {
		pageInput[in.field] = in
	}
	norm := func(typ, id string) string { return strings.ReplaceAll(typ, id, "§") }
	var conv []string
	for _, in := range m.sc.inputs {
		pin, ok := pageInput[in.field]
		if !ok {
			continue
		}
		if norm(pin.typ, ident) != norm(in.typ, masterIdent) {
			return fail(fmt.Sprintf("the input %q is %s on the page and %s on the master", in.name, pin.typ, in.typ))
		}
		conv = append(conv, fmt.Sprintf("%s: p.%s", in.field, in.field))
	}
	if m.sc.palette && sc.palette {
		conv = append(conv, "Palette: master.Palette(p.Palette)")
	}
	// The page's hint and row types that the master has too are the master's.
	masterType := map[string]sharedType{}
	for _, t := range m.sc.shared {
		masterType[strings.TrimPrefix(t.name, masterIdent)] = t
	}
	contract := sc.contract
	for _, t := range sc.shared {
		suffix := strings.TrimPrefix(t.name, ident)
		mt, ok := masterType[suffix]
		if !ok || mt.field != t.field || !strings.HasPrefix(t.name, ident) {
			continue
		}
		if mt.hint != t.hint || mt.signature != t.signature {
			return fail(fmt.Sprintf("the type of %s is not the master's", t.field))
		}
		replacement := fmt.Sprintf("// %s is %s of the master.\ntype %s = master.%s\n\n", t.name, t.name, t.name, suffix)
		if t.hint {
			replacement = fmt.Sprintf("// %s is one hint of the key bar %s, the master's.\ntype %s = master.%s\n\n", t.name, t.field, t.name, suffix)
			replacement += fmt.Sprintf("// %sAt is the index of the hint of %s under column x of the bar (X of the\n// click's event), and false between hints.\nfunc %sAt(p %sProps, x int) (int, bool) {\n\treturn master.%sAt(%sMasterProps(p), x)\n}\n\n",
				t.name, t.field, t.name, ident, suffix, camel(doc.Name))
		}
		whole := t.decl + t.at
		if !strings.Contains(contract, whole) {
			return fail("its types were not written as expected")
		}
		contract = strings.Replace(contract, whole, replacement, 1)
	}
	// The view: the master's parts are taken from the package.
	lower := camel(doc.Name)
	header := fmt.Sprintf("var %sParts = []screenPart[%sProps]{\n", lower, ident)
	old := header + strings.Join(sc.partTexts[:count], "")
	if !strings.Contains(sc.view, old) {
		return fail("its parts were not written as expected")
	}
	sort.Strings(hidden)
	call := fmt.Sprintf("master.PartsFor(%sMasterProps", lower)
	if len(hidden) > 0 {
		call += ", " + strings.Join(hidden, ", ")
	}
	call += ")"
	view := strings.Replace(sc.view, old, fmt.Sprintf("var %sParts = append(%s, []screenPart[%sProps]{\n", lower, call, ident)+"", 1)
	end := fmt.Sprintf("}\n\n// %s draws the screen", ident)
	if !strings.Contains(view, end) {
		return fail("its view was not written as expected")
	}
	view = strings.Replace(view, end, fmt.Sprintf("}...)\n\n// %sMasterProps is the props of the master's parts, from the screen's.\nfunc %sMasterProps(p %sProps) master.Props {\n\treturn master.Props{%s}\n}\n\n// %s draws the screen", lower, lower, ident, strings.Join(conv, ", "), ident), 1)
	if !strings.Contains(view, "strconv.") {
		view = strings.Replace(view, "\t\"strconv\"\n\n", "", 1)
	}
	sc.view, sc.contract = view, contract
	return sc, ""
}
