package expr

import "testing"

func envOf(inputs map[string]float64, rects map[string][4]int, w, h int) *Env {
	return &Env{
		W: w, H: h,
		Input: func(name string) (float64, bool) { v, ok := inputs[name]; return v, ok },
		Rect: func(name, field string) (float64, bool) {
			r, ok := rects[name]
			if !ok {
				return 0, false
			}
			return RectField(r[0], r[1], r[2], r[3], field), true
		},
	}
}

func TestExpressionsReadInputsAndPlacesOfOtherComponents(t *testing.T) {
	env := envOf(map[string]float64{"PlaylistCount": 7, "Open": 1}, map[string][4]int{"Playlists": {0, 4, 40, 9}}, 120, 36)
	cases := []struct {
		src    string
		parent int
		want   int
	}{
		{"min(max($PlaylistCount + 2, 5), (100% - 4) / 3)", 36, 9},
		{"max($PlaylistCount + 2, 5)", 36, 9},
		{`below("Playlists")`, 36, 13},
		{`100% - 4 - height("Playlists")`, 36, 23},
		{`top(Playlists) + $Open`, 36, 5},
		{`$"Missing input" + 3`, 36, 3},
		{`right("Nothing")`, 36, 0},
	}
	for _, c := range cases {
		e, err := Parse(c.src)
		if err != nil {
			t.Fatalf("%q: %v", c.src, err)
		}
		if got := e.ResolveIn(c.parent, env); got != c.want {
			t.Errorf("%q = %d, want %d", c.src, got, c.want)
		}
		if !e.UsesEnv() {
			t.Errorf("%q does not report that it reads the environment", c.src)
		}
		if e.IsFixed() {
			t.Errorf("%q is reported fixed", c.src)
		}
	}
	e, _ := Parse("min($A, $B) + below(\"X\")")
	in, comp := e.Refs()
	if len(in) != 2 || in[0] != "A" || len(comp) != 1 || comp[0] != "X" {
		t.Errorf("refs %v %v", in, comp)
	}
}

func TestADragLeavesAnExpressionThatReadsInputsAlone(t *testing.T) {
	e, _ := Parse("$Count + 2")
	if got := e.Shift(3, 40); got.String() != e.String() {
		t.Errorf("shifted to %q", got.String())
	}
}

func TestConditions(t *testing.T) {
	env := envOf(map[string]float64{"Count": 3, "Busy": 0, "Compact": 1}, nil, 120, 36)
	cases := []struct {
		src  string
		want bool
	}{
		{"w >= 100", true},
		{"w < 100", false},
		{"h < 30 || $Compact", true},
		{"$Count > 0 && !$Busy", true},
		{"$Count > 5 || $Busy", false},
		{"($Count > 0 && w > 200) || $Compact", true},
		{"$Count", true},
		{"$Busy", false},
		{"!($Count == 3)", false},
		{"$Count + 2 == 5", true},
		{"w != 120", false},
	}
	for _, c := range cases {
		cond, err := ParseCond(c.src)
		if err != nil {
			t.Fatalf("%q: %v", c.src, err)
		}
		if got := cond.Eval(env); got != c.want {
			t.Errorf("%q = %v, want %v", c.src, got, c.want)
		}
	}
	for _, bad := range []string{"", "w >=", "&& $A", `below("X") > 1`, "w >> 3"} {
		if _, err := ParseCond(bad); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	if !IsCondition("w >= 100") || IsCondition("Show status") || !IsCondition("!$Busy") {
		t.Error("IsCondition tells names from conditions wrongly")
	}
}
