package expr

import (
	"errors"
	"fmt"
	"strings"
)

// Cond is a condition: what decides whether a component is shown. It compares
// numbers made of the window (w, h), inputs ($Count) and calculations:
// "w >= 100", "$Count > 0 && !$Busy", "h < 30 || $Compact".
type Cond struct {
	root condNode
	src  string
}

type condNode interface {
	test(env *Env) bool
	walk(inputs, components *[]string)
}

// IsCondition reports whether text is written as a condition (it has an
// operator) rather than as the name of an input.
func IsCondition(text string) bool {
	return strings.ContainsAny(text, "<>=!&|$()")
}

// ParseCond reads a condition. A number or an input on its own is true when
// it is not 0.
func ParseCond(s string) (Cond, error) {
	if strings.TrimSpace(s) == "" {
		return Cond{}, ErrEmpty
	}
	p := &parser{src: s}
	root, err := p.or()
	if err != nil {
		return Cond{}, err
	}
	p.skip()
	if p.pos < len(p.src) {
		return Cond{}, fmt.Errorf("unexpected %q at %d", p.src[p.pos:p.pos+1], p.pos+1)
	}
	if hasRect(root) {
		return Cond{}, errors.New("a condition cannot read the place of a component")
	}
	return Cond{root: root, src: clean(s)}, nil
}

// Eval is the condition for the window and inputs in env.
func (c Cond) Eval(env *Env) bool { return c.root == nil || c.root.test(env) }

// Refs lists the inputs the condition reads.
func (c Cond) Refs() []string {
	var inputs, components []string
	if c.root != nil {
		c.root.walk(&inputs, &components)
	}
	return inputs
}

// String is the condition as written, trimmed.
func (c Cond) String() string { return c.src }

func hasRect(n condNode) bool {
	var in, comp []string
	n.walk(&in, &comp)
	return len(comp) > 0
}

type condOr struct{ l, r condNode }

func (n condOr) test(env *Env) bool { return n.l.test(env) || n.r.test(env) }
func (n condOr) walk(i, c *[]string) {
	n.l.walk(i, c)
	n.r.walk(i, c)
}

type condAnd struct{ l, r condNode }

func (n condAnd) test(env *Env) bool { return n.l.test(env) && n.r.test(env) }
func (n condAnd) walk(i, c *[]string) {
	n.l.walk(i, c)
	n.r.walk(i, c)
}

type condNot struct{ x condNode }

func (n condNot) test(env *Env) bool  { return !n.x.test(env) }
func (n condNot) walk(i, c *[]string) { n.x.walk(i, c) }

// condCmp compares two numbers; with no operator it tests l against 0.
type condCmp struct {
	op   string
	l, r node
}

func (n condCmp) test(env *Env) bool {
	l := n.l.eval(0, env)
	if n.op == "" {
		return l != 0
	}
	r := n.r.eval(0, env)
	switch n.op {
	case "==":
		return l == r
	case "!=":
		return l != r
	case ">=":
		return l >= r
	case "<=":
		return l <= r
	case ">":
		return l > r
	}
	return l < r
}

func (n condCmp) walk(inputs, components *[]string) {
	walk(n.l, inputs, components)
	if n.r != nil {
		walk(n.r, inputs, components)
	}
}

func (p *parser) or() (condNode, error) {
	l, err := p.and()
	if err != nil {
		return nil, err
	}
	for p.lookingAt("||") {
		p.pos += 2
		r, err := p.and()
		if err != nil {
			return nil, err
		}
		l = condOr{l, r}
	}
	return l, nil
}

func (p *parser) and() (condNode, error) {
	l, err := p.not()
	if err != nil {
		return nil, err
	}
	for p.lookingAt("&&") {
		p.pos += 2
		r, err := p.not()
		if err != nil {
			return nil, err
		}
		l = condAnd{l, r}
	}
	return l, nil
}

func (p *parser) not() (condNode, error) {
	if p.peek() == '!' && !p.lookingAt("!=") {
		p.pos++
		x, err := p.not()
		if err != nil {
			return nil, err
		}
		return condNot{x}, nil
	}
	// An arithmetic side, then maybe a comparison.
	save := p.pos
	l, err := p.sum()
	if err == nil {
		op := p.comparison()
		if op == "" {
			return condCmp{l: l}, nil
		}
		r, err := p.sum()
		if err != nil {
			return nil, err
		}
		return condCmp{op: op, l: l, r: r}, nil
	}
	// Not arithmetic: a group of conditions.
	p.pos = save
	if p.peek() == '(' {
		p.pos++
		x, gerr := p.or()
		if gerr != nil {
			return nil, gerr
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("missing ) at %d", p.pos+1)
		}
		p.pos++
		return x, nil
	}
	return nil, err
}

func (p *parser) lookingAt(s string) bool {
	p.skip()
	return strings.HasPrefix(p.src[p.pos:], s)
}

// comparison consumes a comparison operator and returns it, or "".
func (p *parser) comparison() string {
	p.skip()
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if strings.HasPrefix(p.src[p.pos:], op) {
			p.pos += len(op)
			return op
		}
	}
	return ""
}
