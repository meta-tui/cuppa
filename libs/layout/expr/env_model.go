package expr

import (
	"math"
)

// Env is what an expression can read besides the size of its parent: the
// inputs of a screen ($Name) and the places of the components before it.
type Env struct {
	// Input is the value of an input; a yes/no input is 1 or 0.
	Input func(name string) (float64, bool)
	// Rect is one edge of a component placed before: field is "x", "y", "w",
	// "h", "right" or "bottom".
	Rect func(name, field string) (float64, bool)
	// W and H are the size of the window, for conditions.
	W, H int
}

// inputRef is $Name.
type inputRef struct{ name string }

func (n inputRef) eval(_ float64, env *Env) float64 {
	if env == nil || env.Input == nil {
		return 0
	}
	v, _ := env.Input(n.name)
	return v
}
func (inputRef) usesParent() bool { return false }
func (inputRef) usesEnv() bool    { return true }

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

// windowDim is w or h: the size of the window in a condition.
type windowDim struct{ axis byte }

func (n windowDim) eval(parent float64, env *Env) float64 {
	if env != nil {
		if n.axis == 'w' {
			return float64(env.W)
		}
		return float64(env.H)
	}
	return parent
}
func (windowDim) usesParent() bool { return false }
func (windowDim) usesEnv() bool    { return true }

// Refs lists the inputs ($Name) and the component names an expression reads.
func (e Expr) Refs() (inputs, components []string) {
	walk(e.root, &inputs, &components)
	return inputs, components
}

func walk(n node, inputs, components *[]string) {
	switch v := n.(type) {
	case inputRef:
		*inputs = append(*inputs, v.name)
	case rectRef:
		*components = append(*components, v.name)
	case negation:
		walk(v.x, inputs, components)
	case binary:
		walk(v.l, inputs, components)
		walk(v.r, inputs, components)
	case call:
		for _, a := range v.args {
			walk(a, inputs, components)
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
