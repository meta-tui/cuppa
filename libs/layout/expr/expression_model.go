// Package expr is a size expression: a number of cells, a percentage of the
// parent, or a calculation over them. "10", "50%", "100% - 10" and
// "min(50%, 40)" are expressions.
package expr

import (
	"math"
	"strconv"
	"strings"
)

// Expr is a parsed expression. The zero value is the constant 0.
type Expr struct {
	root node
	src  string
}

// Resolve returns the expression in whole cells for a parent that is parent
// cells long on the same axis. The result is rounded down; it is not clamped,
// so a calculation can come out negative. A division by zero gives 0.
func (e Expr) Resolve(parent int) int { return e.ResolveIn(parent, nil) }

// ResolveIn is Resolve for an expression that also reads inputs and the
// places of other components; env says what they are. A name env does not know
// is 0.
func (e Expr) ResolveIn(parent int, env *Env) int {
	if e.root == nil {
		return 0
	}
	v := e.root.eval(float64(parent), env)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return int(math.Floor(v + 1e-9))
}

// IsFixed reports whether the expression does not depend on the parent size.
func (e Expr) IsFixed() bool { return e.root == nil || !e.root.usesParent() && !e.root.usesEnv() }

// UsesEnv reports whether the expression reads an input or the place of
// another component.
func (e Expr) UsesEnv() bool { return e.root != nil && e.root.usesEnv() }

// String is the expression as written, trimmed.
func (e Expr) String() string {
	if e.root == nil {
		return "0"
	}
	return e.src
}

// Cells is the expression for a fixed number of cells.
func Cells(n int) Expr {
	return Expr{root: number{v: float64(n)}, src: strconv.Itoa(n)}
}

// Percent is the expression for p percent of the parent.
func Percent(p float64) Expr {
	s := strconv.FormatFloat(p, 'f', -1, 64) + "%"
	return Expr{root: percent{v: p}, src: s}
}

// node is one part of the expression tree.
type node interface {
	eval(parent float64, env *Env) float64
	usesParent() bool
	// usesEnv reports whether the node reads an input or a component's place.
	usesEnv() bool
}

type number struct{ v float64 }

func (n number) eval(float64, *Env) float64 { return n.v }
func (number) usesParent() bool             { return false }
func (number) usesEnv() bool                { return false }

type percent struct{ v float64 }

func (p percent) eval(parent float64, _ *Env) float64 { return parent * p.v / 100 }
func (percent) usesParent() bool                     { return true }
func (percent) usesEnv() bool                        { return false }

type negation struct{ x node }

func (n negation) eval(parent float64, env *Env) float64 { return -n.x.eval(parent, env) }
func (n negation) usesParent() bool                      { return n.x.usesParent() }
func (n negation) usesEnv() bool                         { return n.x.usesEnv() }

type binary struct {
	op   byte
	l, r node
}

func (b binary) eval(parent float64, env *Env) float64 {
	l, r := b.l.eval(parent, env), b.r.eval(parent, env)
	switch b.op {
	case '+':
		return l + r
	case '-':
		return l - r
	case '*':
		return l * r
	}
	if r == 0 {
		return 0
	}
	return l / r
}
func (b binary) usesParent() bool { return b.l.usesParent() || b.r.usesParent() }
func (b binary) usesEnv() bool    { return b.l.usesEnv() || b.r.usesEnv() }

type call struct {
	fn   string
	args []node
}

func (c call) eval(parent float64, env *Env) float64 {
	out := c.args[0].eval(parent, env)
	for _, a := range c.args[1:] {
		v := a.eval(parent, env)
		if c.fn == "min" {
			out = math.Min(out, v)
		} else {
			out = math.Max(out, v)
		}
	}
	return out
}

func (c call) usesParent() bool {
	for _, a := range c.args {
		if a.usesParent() {
			return true
		}
	}
	return false
}

func (c call) usesEnv() bool {
	for _, a := range c.args {
		if a.usesEnv() {
			return true
		}
	}
	return false
}

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }
