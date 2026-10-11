package editor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/layout/expr"
)

// EmptyDefault is the value that says an input starts empty in the exported
// Default<Screen>Props: nil, false, 0 or "" by its type.
const EmptyDefault = "(empty)"

// InputNames are the inputs of the shown page: the ones a property or a
// visibility is bound to and the ones an expression reads, in name order.
func (e *Editor) InputNames() []string {
	seen := map[string]bool{}
	var scan func(nodes []design.Node)
	scan = func(nodes []design.Node) {
		for _, n := range nodes {
			for _, name := range n.Bind {
				seen[name] = true
			}
			if n.ShowIf != "" && !expr.IsCondition(n.ShowIf) && design.ValidInputName(n.ShowIf) {
				seen[n.ShowIf] = true
			}
			scan(n.Children)
		}
	}
	scan(e.doc.Nodes)
	for _, name := range e.ReadInputs() {
		seen[name] = true
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Defaults are the app defaults of the shown page, input name to text.
func (e *Editor) Defaults() map[string]string {
	out := make(map[string]string, len(e.doc.Defaults))
	for k, v := range e.doc.Defaults {
		out[k] = v
	}
	return out
}

// SetDefault gives an input the value the exported Default<Screen>Props starts
// from, as one undo step; an empty value takes it away, and EmptyDefault makes
// the input start empty. The designer still previews the design's own value.
func (e *Editor) SetDefault(name, value string) error {
	value = strings.TrimSpace(value)
	if !design.ValidInputName(name) {
		return fmt.Errorf("editor: %q is not an input name", name)
	}
	if len(value) > 200 || strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("editor: a default is one line of at most 200 characters")
	}
	if e.doc.Defaults[name] == value {
		return nil
	}
	e.apply(func() bool {
		defaults := e.Defaults()
		if value == "" {
			delete(defaults, name)
		} else {
			defaults[name] = value
		}
		if len(defaults) == 0 {
			defaults = nil
		}
		e.doc.Defaults = defaults
		return true
	})
	return nil
}
