package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/menuvalue"
)

var (
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	identifier  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

var kinds = map[string]bool{"text": true, "int": true, "float": true, "bool": true, "color": true, "choice": true, "menu": true}

var families = map[string]bool{"lipgloss": true, "bubbles": true, "huh": true, "glamour": true, "ntcharts": true, "community": true}

// Decode reads a description. An error means the file is not JSON of the
// right shape; issues are the problems of a readable file, in plain words.
// Unknown fields are an issue too, so a misspelt property name is not lost.
func Decode(data []byte) (d Description, issues []string, err error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&d); err != nil {
		return Description{}, nil, fmt.Errorf("component: %w", err)
	}
	strict := json.NewDecoder(bytes.NewReader(data))
	strict.DisallowUnknownFields()
	var again Description
	if err := strict.Decode(&again); err != nil {
		issues = append(issues, "unknown or misplaced field: "+err.Error())
	}
	return d, append(issues, d.Check()...), nil
}

// Check lists what is wrong with the description.
func (d Description) Check() []string {
	var out []string
	bad := func(format string, a ...any) { out = append(out, fmt.Sprintf(format, a...)) }

	if d.Cuppa != Version {
		bad("cuppa is %d; this reader understands %d", d.Cuppa, Version)
	}
	if !namePattern.MatchString(d.Name) {
		bad("name %q must be lowercase letters, digits and single dashes", d.Name)
	}
	if strings.TrimSpace(d.Title) == "" {
		bad("title is empty")
	}
	if strings.TrimSpace(d.Description) == "" {
		bad("description is empty")
	}
	if strings.TrimSpace(d.License) == "" {
		bad("license is empty: a generated program imports the code, so its licence must be known")
	}
	if d.Family != "" && !families[d.Family] {
		bad("family %q is not one of lipgloss, bubbles, huh, glamour, ntcharts, community", d.Family)
	}
	if d.Go.Import == "" {
		bad("go.import is empty")
	}
	if d.Go.Module == "" {
		bad("go.module is empty")
	}
	if d.Go.BubbleTea != 1 && d.Go.BubbleTea != 2 {
		bad("go.bubbletea is %d; it must be 1 or 2", d.Go.BubbleTea)
	}
	if d.Go.Model == "" || d.Go.Constructor == "" {
		bad("go.model and go.constructor are needed to create the component")
	}
	for _, name := range []struct{ field, v string }{
		{"go.model", d.Go.Model}, {"go.constructor", d.Go.Constructor}, {"go.init", d.Go.Init}, {"go.update", d.Go.Update},
		{"go.view", d.Go.View}, {"go.width", d.Go.Width}, {"go.height", d.Go.Height}, {"go.origin", d.Go.Origin},
	} {
		if name.v != "" && !identifier.MatchString(name.v) {
			bad("%s %q is not a Go identifier", name.field, name.v)
		}
	}
	if d.Size.Default.W < 1 || d.Size.Default.H < 1 {
		bad("size.default must be at least 1 by 1")
	}
	if d.Size.Min.W < 1 || d.Size.Min.H < 1 {
		bad("size.min must be at least 1 by 1")
	}
	if d.Size.Min.W > d.Size.Default.W || d.Size.Min.H > d.Size.Default.H {
		bad("size.min is larger than size.default")
	}

	seen := map[string]bool{}
	for i, p := range d.Props {
		where := fmt.Sprintf("props[%d] (%s)", i, p.Key)
		switch {
		case !identifier.MatchString(p.Key):
			bad("%s: key must be letters, digits and underscores", where)
		case seen[p.Key]:
			bad("%s: the key is used twice", where)
		}
		seen[p.Key] = true
		if strings.TrimSpace(p.Label) == "" {
			bad("%s: label is empty", where)
		}
		if !kinds[p.Kind] {
			bad("%s: kind %q is not one of text, int, float, bool, color, choice, menu", where, p.Kind)
			continue
		}
		switch p.Kind {
		case "choice":
			if len(p.Choices) < 2 {
				bad("%s: a choice needs at least two choices", where)
			}
			if !contains(p.Choices, p.Default) {
				bad("%s: default %q is not one of the choices", where, p.Default)
			}
		case "int":
			n, err := strconv.Atoi(p.Default)
			switch {
			case err != nil:
				bad("%s: default %q is not a number", where, p.Default)
			case p.Max > p.Min && (n < p.Min || n > p.Max):
				bad("%s: default %d is outside %d to %d", where, n, p.Min, p.Max)
			}
		case "bool":
			if p.Default != "true" && p.Default != "false" {
				bad("%s: default %q must be true or false", where, p.Default)
			}
		case "menu":
			if _, err := menuvalue.Parse(p.Default); err != nil {
				bad("%s: default is not a list of menus: %v", where, err)
			}
		}
		for _, name := range []struct{ field, v string }{
			{"go.option", p.Go.Option}, {"go.setter", p.Go.Setter}, {"go.getter", p.Go.Getter},
		} {
			if name.v != "" && !identifier.MatchString(name.v) {
				bad("%s: %s %q is not a Go identifier", where, name.field, name.v)
			}
		}
		if p.Go.Option == "" && p.Go.Setter == "" {
			bad("%s: neither go.option nor go.setter says how the property reaches the component", where)
		}
	}

	events := map[string]bool{}
	for i, e := range d.Events {
		where := fmt.Sprintf("events[%d] (%s)", i, e.Name)
		switch {
		case !namePattern.MatchString(e.Name):
			bad("%s: name must be lowercase letters, digits and single dashes", where)
		case events[e.Name]:
			bad("%s: the name is used twice", where)
		}
		events[e.Name] = true
		if !identifier.MatchString(e.Go.Message) {
			bad("%s: go.message %q is not a Go identifier", where, e.Go.Message)
		}
		for _, f := range e.Fields {
			if f.Prop != "" && !seen[f.Prop] {
				bad("%s: field %s carries property %q, which does not exist", where, f.Name, f.Prop)
			}
		}
	}

	switch d.Preview.Kind {
	case "", "frame":
	case "static":
		if len(d.Preview.Lines) == 0 {
			bad("preview is static but has no lines")
		}
	default:
		bad("preview.kind %q is not frame or static", d.Preview.Kind)
	}
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
