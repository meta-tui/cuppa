package expr

import "testing"

func envOf(rects map[string][4]int) *Env {
	return &Env{
		Rect: func(name, field string) (float64, bool) {
			r, ok := rects[name]
			if !ok {
				return 0, false
			}
			return RectField(r[0], r[1], r[2], r[3], field), true
		},
	}
}

func TestExpressionsReadThePlacesOfOtherComponents(t *testing.T) {
	env := envOf(map[string][4]int{"Playlists": {0, 4, 40, 9}})
	cases := []struct {
		src    string
		parent int
		want   int
	}{
		{"min(max(height(Playlists) + 2, 5), (100% - 4) / 3)", 36, 10},
		{`below("Playlists")`, 36, 13},
		{`100% - 4 - height("Playlists")`, 36, 23},
		{`top(Playlists) + 1`, 36, 5},
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
	e, _ := Parse("min(width(\"A\"), 5) + below(\"X\")")
	if refs := e.Refs(); len(refs) != 2 || refs[0] != "A" || refs[1] != "X" {
		t.Errorf("refs %v", refs)
	}
}
