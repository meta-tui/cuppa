package editor

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// The live values of a running application (ADR 0016): the designer draws them
// instead of the values the design holds, and keeps nothing of them in the
// document. They are given by page name and input name, as the application
// publishes them.

// SetLive gives the values the application published for a page. A page's set
// replaces its earlier one.
func (e *Editor) SetLive(page string, inputs map[string]json.RawMessage) {
	if e.live == nil {
		e.live = map[string]map[string]json.RawMessage{}
	}
	e.live[page] = inputs
}

// ClearLive goes back to the values of the design.
func (e *Editor) ClearLive() { e.live = nil }

// Live reports whether the designer is drawing an application's values.
func (e *Editor) Live() bool { return e.live != nil }

// livePageName is the name the shown page is published under.
func (e *Editor) livePageName() string {
	names := e.doc.PageNames()
	return names[min(max(e.doc.ActivePage(), 0), len(names)-1)]
}

// applyLive puts the published values of the shown page into doc: bound
// properties, the numbers layout expressions read and the yes/no inputs that
// show or hide a component.
func (e *Editor) applyLive(doc *design.Document) bool {
	inputs := e.live[e.livePageName()]
	if len(inputs) == 0 {
		return false
	}
	for name, raw := range inputs {
		if !design.ValidInputName(name) {
			continue
		}
		var v any
		if json.Unmarshal(raw, &v) != nil {
			continue
		}
		switch x := v.(type) {
		case float64:
			setSample(doc, name, strconv.FormatFloat(x, 'f', -1, 64))
		case bool:
			setSample(doc, name, strconv.FormatBool(x))
		}
	}
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for i := range nodes {
			n := &nodes[i]
			props := make(map[string]string, len(n.Props)+len(n.Bind))
			for k, v := range n.Props {
				props[k] = v
			}
			for key, input := range n.Bind {
				raw, ok := inputs[input]
				if !ok {
					continue
				}
				text, extra, ok := liveText(*n, key, raw)
				if !ok {
					continue
				}
				props[key] = text
				for k, v := range extra {
					props[k] = v
				}
			}
			n.Props = props
			if n.ShowIf != "" && design.ValidInputName(n.ShowIf) {
				var shown bool
				if raw, ok := inputs[n.ShowIf]; ok && json.Unmarshal(raw, &shown) == nil && !shown {
					n.Hidden = true
				}
			}
			walk(n.Children)
		}
	}
	walk(doc.Nodes)
	return true
}

func setSample(doc *design.Document, name, value string) {
	samples := make(map[string]string, len(doc.Samples)+1)
	for k, v := range doc.Samples {
		samples[k] = v
	}
	samples[name] = value
	doc.Samples = samples
}

// liveText is the property text of a published value, and the properties that
// go with it (the styles of rows). ok is false for a value the property cannot
// show.
func liveText(n design.Node, key string, raw json.RawMessage) (text string, extra map[string]string, ok bool) {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return "", nil, false
	}
	switch x := v.(type) {
	case string:
		return x, nil, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil, true
	case bool:
		return strconv.FormatBool(x), nil, true
	case []any:
		return liveList(n, x)
	}
	return "", nil, false
}

// liveList is a list: strings, numbers (open sections), key bar hints
// ({Key, Label}) or rows (objects with a field per column).
func liveList(n design.Node, items []any) (string, map[string]string, bool) {
	if len(items) == 0 {
		return "", map[string]string{"styles": "", "cellstyles": ""}, true
	}
	switch items[0].(type) {
	case string, float64:
		cells := make([]string, len(items))
		for i, it := range items {
			switch x := it.(type) {
			case string:
				cells[i] = liveEscape(x)
			case float64:
				cells[i] = strconv.FormatFloat(x, 'f', -1, 64)
			default:
				return "", nil, false
			}
		}
		return strings.Join(cells, ","), nil, true
	case map[string]any:
		objs := make([]map[string]any, len(items))
		for i, it := range items {
			o, ok := it.(map[string]any)
			if !ok {
				return "", nil, false
			}
			objs[i] = o
		}
		if _, hint := objs[0]["Key"]; hint {
			cells := make([]string, len(objs))
			for i, o := range objs {
				cells[i] = liveEscape(liveCell(o, "Key") + ":" + liveCell(o, "Label"))
			}
			return strings.Join(cells, ","), nil, true
		}
		columns := splitColumns(n.Props["columns"])
		rows := make([]string, len(objs))
		styles := make([]string, len(objs))
		for i, o := range objs {
			cells := make([]string, len(columns))
			for j, c := range columns {
				cells[j] = liveEscape(liveCell(o, c))
			}
			rows[i] = strings.Join(cells, ",")
			styles[i] = "normal"
			if s := liveCell(o, "Style"); s != "" {
				styles[i] = s
			}
		}
		return strings.Join(rows, ";"), map[string]string{"styles": strings.Join(styles, ","), "cellstyles": ""}, true
	}
	return "", nil, false
}

// splitColumns are the column names of a Rows "columns" property
// ("Mark:2,Name:14,Swatch:6:colour,Hex").
func splitColumns(spec string) []string {
	var out []string
	for _, item := range strings.Split(spec, ",") {
		name, _, _ := strings.Cut(strings.TrimSpace(item), ":")
		out = append(out, strings.TrimSpace(name))
	}
	return out
}

// liveCell is the text of the field of o that a column names: the same name
// without regard to case, spaces or punctuation.
func liveCell(o map[string]any, column string) string {
	want := squash(column)
	for k, v := range o {
		if squash(k) != want {
			continue
		}
		switch x := v.(type) {
		case string:
			return x
		case float64:
			return strconv.FormatFloat(x, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(x)
		}
	}
	return ""
}

func squash(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

var liveEscaper = strings.NewReplacer(`\`, `\\`, ",", `\,`, ";", `\;`)

func liveEscape(s string) string { return liveEscaper.Replace(s) }
