package editor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/layout/expr"
)

// ReadInputs are the inputs that a layout expression or a condition reads and
// no property is bound to: the ones whose value only a sample can give in the
// designer, in name order.
func (e *Editor) ReadInputs() []string {
	bound := map[string]bool{}
	read := map[string]bool{}
	var scan func(nodes []design.Node)
	scan = func(nodes []design.Node) {
		for _, n := range nodes {
			for _, name := range n.Bind {
				bound[name] = true
			}
			for _, src := range []string{n.Layout.X, n.Layout.Y, n.Layout.W, n.Layout.H} {
				if src == "" {
					continue
				}
				if x, err := expr.Parse(src); err == nil {
					refs, _ := x.Refs()
					for _, r := range refs {
						read[r] = true
					}
				}
			}
			if n.ShowIf != "" && expr.IsCondition(n.ShowIf) {
				if c, err := expr.ParseCond(n.ShowIf); err == nil {
					for _, r := range c.Refs() {
						read[r] = true
					}
				}
			}
			scan(n.Children)
		}
	}
	scan(e.doc.Nodes)
	for _, p := range e.doc.Pages {
		scan(p.Nodes)
	}
	var out []string
	for name := range read {
		if !bound[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// Samples are the sample values of the design, input name to number.
func (e *Editor) Samples() map[string]string {
	out := make(map[string]string, len(e.doc.Samples))
	for k, v := range e.doc.Samples {
		out[k] = v
	}
	return out
}

// SetSample gives an input the value the designer previews it with, as one
// undo step; an empty value takes the sample away. The value is a number, or
// true and false.
func (e *Editor) SetSample(name, value string) error {
	value = strings.TrimSpace(value)
	if !design.ValidInputName(name) {
		return fmt.Errorf("editor: %q is not an input name", name)
	}
	if value != "" {
		if _, ok := design.SampleNumber(value); !ok {
			return fmt.Errorf("editor: a sample is a number, or true or false")
		}
	}
	if e.doc.Samples[name] == value {
		return nil
	}
	e.apply(func() bool {
		samples := make(map[string]string, len(e.doc.Samples)+1)
		for k, v := range e.doc.Samples {
			samples[k] = v
		}
		if value == "" {
			delete(samples, name)
		} else {
			samples[name] = value
		}
		if len(samples) == 0 {
			samples = nil
		}
		e.doc.Samples = samples
		// Layouts that read the input follow.
		e.Resolve()
		return true
	})
	return nil
}
