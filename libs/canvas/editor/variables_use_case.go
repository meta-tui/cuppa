package editor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/layout/expr"
)

// VariableKind is which namespace a variable lives in: a named colour, a
// screen input (bound property or show-if) or a screen event.
type VariableKind string

// Variable kinds.
const (
	VarColour VariableKind = "colour"
	VarInput  VariableKind = "input"
	VarEvent  VariableKind = "event"
)

// Use is one place a variable is used.
type Use struct {
	// Node is the component, empty for a screen key.
	Node design.NodeID
	// NodeName is the component's name, or "Screen keys".
	NodeName string
	// Where says what of the component uses it: a property label, "Show if",
	// "On click", or the key.
	Where string
	// Prop is the property key of a colour or bound property, or the key of a
	// screen key.
	Prop string
	// Role is what the use is: UseProp, UseShowIf, UseEvent or UseKey.
	Role string
}

// Variable is a named value of the design and where it is used.
type Variable struct {
	Kind VariableKind
	Name string
	// Type is how the variable reads: colour, text, number, yes/no, list,
	// rows or event.
	Type string
	// Value is the colour, or the design's value of an input ("" for events).
	Value string
	Uses  []Use
}

// Use roles.
const (
	UseProp   = "prop"
	UseShowIf = "showif"
	UseEvent  = "event"
	UseKey    = "key"
	// UseLayout is an input read by a layout expression: Prop is the axis.
	UseLayout = "layout"
)

// Variables lists every named variable of the design: the colours of the
// palette, the screen inputs and the events, each with the places it is used.
// A palette colour with no use is listed too.
func (e *Editor) Variables() []Variable {
	var out []Variable
	index := map[string]int{}
	add := func(kind VariableKind, name, typ, value string) int {
		key := string(kind) + ":" + name
		if i, ok := index[key]; ok {
			return i
		}
		index[key] = len(out)
		out = append(out, Variable{Kind: kind, Name: name, Type: typ, Value: value})
		return len(out) - 1
	}
	for _, s := range e.doc.Theme.Palette {
		add(VarColour, s.Name, "colour", s.Color)
	}
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for _, n := range nodes {
			def, known := e.cat.Get(n.Component)
			label := func(key string) string {
				if spec, ok := def.Prop(key); known && ok {
					return spec.Label
				}
				return key
			}
			for _, key := range n.BoundKeys() {
				typ := "text"
				if spec, ok := def.Prop(key); known && ok {
					typ = propType(spec)
				}
				i := add(VarInput, n.Bind[key], typ, n.Props[key])
				out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: label(key), Prop: key, Role: UseProp})
			}
			if n.ShowIf != "" {
				if expr.IsCondition(n.ShowIf) {
					// A condition reads inputs by $Name.
					if cond, err := expr.ParseCond(n.ShowIf); err == nil {
						for _, ref := range cond.Refs() {
							i := add(VarInput, ref, "number", "")
							out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: "Show if", Role: UseShowIf})
						}
					}
				} else {
					i := add(VarInput, n.ShowIf, "yes/no", "true")
					out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: "Show if", Role: UseShowIf})
				}
			}
			for _, axis := range []struct{ key, src string }{{"x", n.Layout.X}, {"y", n.Layout.Y}, {"w", n.Layout.W}, {"h", n.Layout.H}} {
				if !strings.Contains(axis.src, "$") {
					continue
				}
				if x, err := expr.Parse(axis.src); err == nil {
					refs, _ := x.Refs()
					for _, ref := range refs {
						i := add(VarInput, ref, "number", "")
						out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: "Layout " + axis.key, Prop: axis.key, Role: UseLayout})
					}
				}
			}
			if n.Event != "" {
				i := add(VarEvent, n.Event, "event", "")
				out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: "On click", Role: UseEvent})
			}
			for _, key := range sortedPropKeys(n.Props) {
				name, ok := strings.CutPrefix(n.Props[key], design.TokenPrefix)
				if !ok {
					continue
				}
				if spec, ok := def.Prop(key); !known || !ok || spec.Kind != definition.PropColor {
					continue
				}
				if i, ok := index[string(VarColour)+":"+name]; ok {
					out[i].Uses = append(out[i].Uses, Use{Node: n.ID, NodeName: n.Name, Where: label(key), Prop: key, Role: UseProp})
				}
			}
			walk(n.Children)
		}
	}
	walk(e.doc.Nodes)
	for _, k := range e.doc.Keys {
		i := add(VarEvent, k.Event, "event", "")
		out[i].Uses = append(out[i].Uses, Use{NodeName: "Screen keys", Where: "key " + k.Key, Role: UseKey, Prop: k.Key})
	}
	return out
}

func sortedPropKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func propType(spec definition.PropSpec) string {
	switch spec.Kind {
	case definition.PropInt, definition.PropFloat:
		return "number"
	case definition.PropBool:
		return "yes/no"
	case definition.PropColor:
		return "colour"
	}
	switch spec.Key {
	case "rows":
		return "rows"
	case "items", "options", "entries", "tabs", "buttons", "columns", "labels", "headers", "bindings", "checked", "values":
		return "list"
	}
	return "text"
}

// RenameVariable gives a variable a new name everywhere it is used, as one undo
// step. It refuses a name another variable of the same kind already has; to
// merge two variables, point the uses of one at the other with Relink.
func (e *Editor) RenameVariable(kind VariableKind, from, to string) error {
	to = design.CleanInputName(to)
	if !design.ValidInputName(to) {
		return fmt.Errorf("editor: a name starts with a letter and has letters, digits, spaces, - or _")
	}
	if from == to {
		return nil
	}
	exists := false
	for _, v := range e.Variables() {
		if v.Kind == kind && v.Name == to {
			return fmt.Errorf("editor: there is already a variable called %s", to)
		}
		if v.Kind == kind && v.Name == from {
			exists = true
		}
	}
	if !exists {
		return fmt.Errorf("editor: no variable called %s", from)
	}
	e.apply(func() bool {
		switch kind {
		case VarColour:
			palette := append([]design.Swatch(nil), e.doc.Theme.Palette...)
			for i := range palette {
				if palette[i].Name == from {
					palette[i].Name = to
				}
			}
			e.doc.Theme.Palette = palette
		case VarEvent:
			keys := append([]design.KeyBinding(nil), e.doc.Keys...)
			for i := range keys {
				if keys[i].Event == from {
					keys[i].Event = to
				}
			}
			e.doc.Keys = keys
		}
		renameIn(e.doc.Nodes, kind, from, to)
		return true
	})
	return nil
}

// renameIn rewrites every use of a variable in the nodes, groups included.
func renameIn(nodes []design.Node, kind VariableKind, from, to string) {
	for i := range nodes {
		n := &nodes[i]
		switch kind {
		case VarColour:
			for k, v := range n.Props {
				if v == design.TokenPrefix+from {
					n.Props[k] = design.TokenPrefix + to
				}
			}
		case VarInput:
			for k, v := range n.Bind {
				if v == from {
					n.Bind[k] = to
				}
			}
			if n.ShowIf == from {
				n.ShowIf = to
			} else if expr.IsCondition(n.ShowIf) {
				n.ShowIf = renameRef(n.ShowIf, from, to)
			}
			n.Layout.X, n.Layout.Y = renameRef(n.Layout.X, from, to), renameRef(n.Layout.Y, from, to)
			n.Layout.W, n.Layout.H = renameRef(n.Layout.W, from, to), renameRef(n.Layout.H, from, to)
		case VarEvent:
			if n.Event == from {
				n.Event = to
			}
		}
		renameIn(n.Children, kind, from, to)
	}
}

// Relink points one use at another variable of the same kind. An input or an
// event name need not exist yet; a colour must already be in the palette. One
// undo step.
func (e *Editor) Relink(u Use, kind VariableKind, to string) error {
	to = design.CleanInputName(to)
	if !design.ValidInputName(to) {
		return fmt.Errorf("editor: a name starts with a letter and has letters, digits, spaces, - or _")
	}
	switch {
	case kind == VarColour && u.Role == UseProp:
		return e.UseSwatch(u.Node, u.Prop, to)
	case kind == VarInput && u.Role == UseProp:
		return e.SetBinding(u.Node, u.Prop, to)
	case kind == VarInput && u.Role == UseShowIf:
		return e.SetShowIf(u.Node, to)
	case kind == VarEvent && u.Role == UseEvent:
		return e.SetEvent(u.Node, to)
	case kind == VarEvent && u.Role == UseKey:
		keys := make([]design.KeyBinding, len(e.doc.Keys))
		copy(keys, e.doc.Keys)
		for i := range keys {
			if keys[i].Key == u.Prop {
				keys[i].Event = to
			}
		}
		return e.SetKeys(design.FormatKeys(keys))
	}
	return fmt.Errorf("editor: cannot link that use")
}

// renameRef renames an input in an expression or a condition: $Name and
// $"A name" become $New or $"A new".
func renameRef(text, from, to string) string {
	if !strings.Contains(text, "$") {
		return text
	}
	plain := func(s string) bool {
		for _, r := range s {
			if r != '_' && (r < '0' || r > '9') && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return false
			}
		}
		return s != ""
	}
	repl := "$" + to
	if !plain(to) {
		repl = "$" + strconv.Quote(to)
	}
	text = strings.ReplaceAll(text, "$"+strconv.Quote(from), repl)
	if plain(from) {
		text = regexp.MustCompile(`\$`+regexp.QuoteMeta(from)+`\b`).ReplaceAllLiteralString(text, repl)
	}
	return text
}
