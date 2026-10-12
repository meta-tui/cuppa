package expr

import "testing"

func TestResolve(t *testing.T) {
	cases := []struct {
		src    string
		parent int
		want   int
	}{
		{"10", 80, 10},
		{"10 cols", 80, 10},
		{"2row", 40, 2},
		{"100%", 80, 80},
		{"50%", 81, 40},
		{"100% - 10", 120, 110},
		{"100%-60", 120, 60},
		{"(100% - 10) / 2", 120, 55},
		{"2 * 10 + 5", 0, 25},
		{"-5 + 10", 0, 5},
		{"min(50%, 40)", 120, 40},
		{"min(50%, 40)", 60, 30},
		{"max(10, 25%, 5)", 80, 20},
		{"100% - 200", 80, -120},
		{"10 / 0", 80, 0},
		{"33.3%", 100, 33},
	}
	for _, c := range cases {
		e, err := Parse(c.src)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.src, err)
		}
		if got := e.Resolve(c.parent); got != c.want {
			t.Errorf("%q at %d = %d, want %d", c.src, c.parent, got, c.want)
		}
	}
}

func TestParseRefusesWhatItCannotRead(t *testing.T) {
	for _, src := range []string{"", "  ", "10 +", "(10", "10)", "abc", "min()", "min(1", "5 %% 2", "1..2", "10 px", "3 4"} {
		if _, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) succeeded", src)
		}
	}
}

func TestIsFixedAndString(t *testing.T) {
	for src, fixed := range map[string]bool{"10": true, "2*5": true, "50%": false, "min(10, 20%)": false} {
		e, err := Parse(src)
		if err != nil {
			t.Fatal(err)
		}
		if e.IsFixed() != fixed {
			t.Errorf("IsFixed(%q) = %v", src, e.IsFixed())
		}
	}
	e, _ := Parse("  100%   -  10 ")
	if e.String() != "100% - 10" {
		t.Errorf("String() = %q", e.String())
	}
	if Cells(7).String() != "7" || Percent(25).String() != "25%" || Percent(25).Resolve(80) != 20 {
		t.Error("constructors")
	}
	var zero Expr
	if zero.Resolve(80) != 0 || zero.String() != "0" || !zero.IsFixed() {
		t.Error("zero value is the constant 0")
	}
}
