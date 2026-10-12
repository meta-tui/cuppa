package expr

import (
	"math"
	"strconv"
	"strings"
)

// Shift returns the expression moved by delta cells, for a parent that is
// parent cells long: what a drag writes back so a node keeps its unit. A fixed
// expression becomes the new number of cells, a lone percentage stays a
// percentage, and a calculation ending in "± n" has that n changed (or gains
// one).
func (e Expr) Shift(delta, parent int) Expr {
	if delta == 0 || e.UsesEnv() {
		return e
	}
	if e.IsFixed() {
		return Cells(e.Resolve(parent) + delta)
	}
	switch r := e.root.(type) {
	case percent:
		if parent <= 0 {
			return Cells(e.Resolve(parent) + delta)
		}
		return Percent(math.Round((r.v+100*float64(delta)/float64(parent))*100) / 100)
	case binary:
		if n, ok := r.r.(number); ok && (r.op == '+' || r.op == '-') {
			v := n.v
			if r.op == '-' {
				v = -v
			}
			return rebuild(r.l, v+float64(delta), e)
		}
	}
	return rebuild(e.root, float64(delta), e)
}

// rebuild writes "l + v" (or "l - |v|", or just l for 0) and parses it back.
func rebuild(l node, v float64, fallback Expr) Expr {
	s := format(l, 0)
	switch {
	case v > 0:
		s += " + " + num(v)
	case v < 0:
		s += " - " + num(-v)
	}
	out, err := Parse(s)
	if err != nil {
		return fallback
	}
	return out
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// precedence of a node when written out.
func precedence(n node) int {
	switch b := n.(type) {
	case binary:
		if b.op == '+' || b.op == '-' {
			return 1
		}
		return 2
	case negation:
		return 3
	}
	return 4
}

// format writes a tree back as text; min is the precedence the context needs,
// below which parentheses are added.
func format(n node, min int) string {
	var s string
	switch v := n.(type) {
	case number:
		s = num(v.v)
	case percent:
		s = num(v.v) + "%"
	case negation:
		s = "-" + format(v.x, 3)
	case binary:
		p := precedence(v)
		s = format(v.l, p) + " " + string(v.op) + " " + format(v.r, p+1)
	case rectRef:
		s = v.fn() + "(" + strconv.Quote(v.name) + ")"
	case call:
		args := make([]string, len(v.args))
		for i, a := range v.args {
			args[i] = format(a, 0)
		}
		s = v.fn + "(" + strings.Join(args, ", ") + ")"
	}
	if precedence(n) < min {
		return "(" + s + ")"
	}
	return s
}

// PercentOf is the percentage of parent that resolves to exactly cells: the
// shortest two-decimal one that does. With no parent it is just the cells.
func PercentOf(cells, parent int) Expr {
	if parent <= 0 {
		return Cells(cells)
	}
	base := math.Round(float64(cells)*10000/float64(parent)) / 100
	for _, d := range []float64{0, 0.01, -0.01, 0.02, -0.02} {
		if e := Percent(math.Round((base+d)*100) / 100); e.Resolve(parent) == cells {
			return e
		}
	}
	return Cells(cells)
}
