package gosource

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// InputContract is one input of a page: what a program publishes under its
// name when it uses the live connection (ADR 0016) or reads from the export.
type InputContract struct {
	Name string `json:"name"`
	// Kind is bool, int, number, string, list (of text), ints, table (rows of
	// text), rows (a list of objects, one field per column), hints (Key and
	// Label) or colour.
	Kind string `json:"kind"`
	// Fields are the columns of rows and hints: the keys of each object.
	Fields []string `json:"fields,omitempty"`
	// Go is the field of the exported props.
	Go string `json:"go"`
	// Use says what the input is for.
	Use string `json:"use,omitempty"`
}

// EventContract is one event a page raises.
type EventContract struct {
	Name string `json:"name"`
	// Raised says what raises it.
	Raised string `json:"raised,omitempty"`
}

// PageContract is what a program has to know of a page: its inputs and events.
type PageContract struct {
	Page   string          `json:"page"`
	Inputs []InputContract `json:"inputs"`
	Events []EventContract `json:"events"`
}

// Contract lists, for each page of the design, its inputs (by the names the
// design gives them, with their kind) and its events: the same contract the Go
// export writes as types, for a program in any language.
func Contract(doc design.Document, cat Catalog) []PageContract {
	var out []PageContract
	for _, page := range doc.Screens() {
		placed := leaves(page, cat)
		sc := buildScreen(page, "Page", placed, cat)
		pc := PageContract{Page: page.Name, Inputs: []InputContract{}, Events: []EventContract{}}
		typed := map[string]sharedType{}
		for _, t := range sc.shared {
			typed[t.field] = t
		}
		for _, in := range sc.inputs {
			if strings.HasPrefix(in.typ, "func") {
				continue
			}
			ic := InputContract{Name: in.name, Go: in.field, Use: in.use}
			switch {
			case in.typ == "bool", in.typ == "int", in.typ == "string":
				ic.Kind = in.typ
			case in.typ == "float64":
				ic.Kind = "number"
			case in.typ == "[]string":
				ic.Kind = "list"
			case in.typ == "[]int":
				ic.Kind = "ints"
			case in.typ == "[][]string":
				ic.Kind = "table"
			case in.typ == "ColourPicker":
				ic.Kind = "colour"
			case strings.HasPrefix(in.typ, "[]"):
				t, ok := typed[in.field]
				switch {
				case ok && t.hint:
					ic.Kind, ic.Fields = "hints", []string{"Key", "Label"}
				case ok:
					ic.Kind, ic.Fields = "rows", append(strings.Split(t.signature, ","), "Style")
				default:
					ic.Kind = "list"
				}
			default:
				ic.Kind = in.typ
			}
			pc.Inputs = append(pc.Inputs, ic)
		}
		for _, e := range sc.events {
			pc.Events = append(pc.Events, EventContract{Name: e, Raised: eventSources(e, placed, page.Keys)})
		}
		out = append(out, pc)
	}
	return out
}
