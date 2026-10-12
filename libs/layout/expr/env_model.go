package expr

import (
	"math"
)

// Env is what an expression can read besides the size of its parent: the
// places of the components before it.
type Env struct {
	// Rect is one edge of a component placed before: field is "x", "y", "w",
	// "h", "right" or "bottom".
	Rect func(name, field string) (float64, bool)
}

// rectRef is below("Name"), right("Name"), top, left, height or width.
type rectRef struct{ name, field string }

func (n rectRef) eval(_ float64, env *Env) float64 {
	if env == nil || env.Rect == nil {
		return 0
	}
	v, _ := env.Rect(n.name, n.field)
	return v
}
func (rectRef) usesParent() bool { return false }
func (rectRef) usesEnv() bool    { return true }

// fn is the function the reference was written with.
func (n rectRef) fn() string {
	for name, field := range rectFunctions {
		if field == n.field {
			return name
		}
	}
	return n.field
}

// Refs lists the component names an expression reads.
func (e Expr) Refs() []string {
	var components []string
	walk(e.root, &components)
	return components
}

func walk(n node, components *[]string) {
	switch v := n.(type) {
	case rectRef:
		*components = append(*components, v.name)
	case negation:
		walk(v.x, components)
	case binary:
		walk(v.l, components)
		walk(v.r, components)
	case call:
		for _, a := range v.args {
			walk(a, components)
		}
	}
}

// RectField is the value of one edge of a rectangle for Env.Rect.
func RectField(x, y, w, h int, field string) float64 {
	switch field {
	case "x":
		return float64(x)
	case "y":
		return float64(y)
	case "w":
		return float64(w)
	case "h":
		return float64(h)
	case "right":
		return float64(x + w)
	case "bottom":
		return float64(y + h)
	}
	return math.NaN()
}
