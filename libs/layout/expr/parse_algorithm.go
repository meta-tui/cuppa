package expr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrEmpty is returned for an expression with nothing in it.
var ErrEmpty = errors.New("empty expression")

// Parse reads an expression. Numbers are cells (an optional "col" or "row"
// after one is accepted and ignored), "n%" is a percentage of the parent, and
// the operators are + - * / with parentheses, unary minus, and min() and max()
// of one or more arguments.
func Parse(s string) (Expr, error) {
	if strings.TrimSpace(s) == "" {
		return Expr{}, ErrEmpty
	}
	p := &parser{src: s}
	root, err := p.sum()
	if err != nil {
		return Expr{}, err
	}
	p.skip()
	if p.pos < len(p.src) {
		return Expr{}, fmt.Errorf("unexpected %q at %d", p.src[p.pos:p.pos+1], p.pos+1)
	}
	return Expr{root: root, src: clean(s)}, nil
}

type parser struct {
	src string
	pos int
}

func (p *parser) skip() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *parser) peek() byte {
	p.skip()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) sum() (node, error) {
	l, err := p.product()
	if err != nil {
		return nil, err
	}
	for c := p.peek(); c == '+' || c == '-'; c = p.peek() {
		p.pos++
		r, err := p.product()
		if err != nil {
			return nil, err
		}
		l = binary{op: c, l: l, r: r}
	}
	return l, nil
}

func (p *parser) product() (node, error) {
	l, err := p.unary()
	if err != nil {
		return nil, err
	}
	for c := p.peek(); c == '*' || c == '/'; c = p.peek() {
		p.pos++
		r, err := p.unary()
		if err != nil {
			return nil, err
		}
		l = binary{op: c, l: l, r: r}
	}
	return l, nil
}

func (p *parser) unary() (node, error) {
	switch p.peek() {
	case '-':
		p.pos++
		x, err := p.unary()
		if err != nil {
			return nil, err
		}
		return negation{x: x}, nil
	case '+':
		p.pos++
		return p.unary()
	}
	return p.atom()
}

func (p *parser) atom() (node, error) {
	switch c := p.peek(); {
	case c == 0:
		return nil, errors.New("expression ends too early")
	case c == '(':
		p.pos++
		x, err := p.sum()
		if err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("missing ) at %d", p.pos+1)
		}
		p.pos++
		return x, nil
	case c >= '0' && c <= '9' || c == '.':
		return p.quantity()
	case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
		return p.function()
	default:
		return nil, fmt.Errorf("unexpected %q at %d", string(c), p.pos+1)
	}
}

func (p *parser) quantity() (node, error) {
	start := p.pos
	for p.pos < len(p.src) && (p.src[p.pos] >= '0' && p.src[p.pos] <= '9' || p.src[p.pos] == '.') {
		p.pos++
	}
	v, err := strconv.ParseFloat(p.src[start:p.pos], 64)
	if err != nil {
		return nil, fmt.Errorf("bad number %q at %d", p.src[start:p.pos], start+1)
	}
	if p.pos < len(p.src) && p.src[p.pos] == '%' {
		p.pos++
		return percent{v: v}, nil
	}
	p.skip()
	rest := p.src[p.pos:]
	for _, unit := range []string{"cols", "col", "rows", "row"} {
		if strings.HasPrefix(rest, unit) && !letterAfter(rest, len(unit)) {
			p.pos += len(unit)
			break
		}
	}
	return number{v: v}, nil
}

func letterAfter(s string, i int) bool {
	return i < len(s) && (s[i] >= 'a' && s[i] <= 'z' || s[i] >= 'A' && s[i] <= 'Z')
}

func (p *parser) function() (node, error) {
	start := p.pos
	for p.pos < len(p.src) && letterAfter(p.src, p.pos) {
		p.pos++
	}
	name := strings.ToLower(p.src[start:p.pos])
	if field, ok := rectFunctions[name]; ok {
		return p.rectRef(name, field)
	}
	if name != "min" && name != "max" {
		return nil, fmt.Errorf("unknown name %q at %d", p.src[start:p.pos], start+1)
	}
	if p.peek() != '(' {
		return nil, fmt.Errorf("%s needs ( at %d", name, p.pos+1)
	}
	p.pos++
	var args []node
	for {
		a, err := p.sum()
		if err != nil {
			return nil, err
		}
		args = append(args, a)
		switch p.peek() {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return call{fn: name, args: args}, nil
		default:
			return nil, fmt.Errorf("expected , or ) at %d", p.pos+1)
		}
	}
}

// rectFunctions are the functions that read the place of a component, with
// the edge of its rectangle each one gives.
var rectFunctions = map[string]string{
	"below": "bottom", "right": "right", "top": "y", "left": "x", "height": "h", "width": "w",
}

// rectRef reads below("Name"), right("Name"), top, left, height or width: one
// edge of a component placed before this one.
func (p *parser) rectRef(fn, field string) (node, error) {
	if p.peek() != '(' {
		return nil, fmt.Errorf("%s needs ( at %d", fn, p.pos+1)
	}
	p.pos++
	var name string
	if p.peek() == '"' {
		q, err := p.quoted()
		if err != nil {
			return nil, err
		}
		name = q
	} else {
		start := p.pos
		for p.pos < len(p.src) && p.src[p.pos] != ')' {
			p.pos++
		}
		name = strings.TrimSpace(p.src[start:p.pos])
	}
	if name == "" {
		return nil, fmt.Errorf("%s needs the name of a component at %d", fn, p.pos+1)
	}
	if p.peek() != ')' {
		return nil, fmt.Errorf("missing ) at %d", p.pos+1)
	}
	p.pos++
	return rectRef{name: name, field: field}, nil
}

// quoted reads a "string" at the current position.
func (p *parser) quoted() (string, error) {
	p.skip()
	if p.pos >= len(p.src) || p.src[p.pos] != '"' {
		return "", fmt.Errorf("a quoted name was expected at %d", p.pos+1)
	}
	p.pos++
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '"' {
		p.pos++
	}
	if p.pos >= len(p.src) {
		return "", fmt.Errorf("the quote opened at %d is not closed", start)
	}
	name := p.src[start:p.pos]
	p.pos++
	return name, nil
}
