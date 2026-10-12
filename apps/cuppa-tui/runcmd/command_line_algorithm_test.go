package runcmd

import (
	"slices"
	"testing"
)

func TestACommandLineSplitsIntoTheProgramAndItsArguments(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`go build -o app ./cmd/app`, []string{"go", "build", "-o", "app", "./cmd/app"}},
		{`"C:/Program Files/app.exe" --demo`, []string{"C:/Program Files/app.exe", "--demo"}},
		{`app "" x`, []string{"app", "", "x"}},
		{`app "a \"b\" c"`, []string{"app", `a "b" c`}},
		{`  app   `, []string{"app"}},
		{``, nil},
	}
	for _, c := range cases {
		got, err := splitCommand(c.in)
		if err != nil || !slices.Equal(got, c.want) {
			t.Errorf("%q: got %q, %v", c.in, got, err)
		}
	}
}

func TestAnUnclosedQuoteIsAnError(t *testing.T) {
	if _, err := splitCommand(`app "x`); err == nil {
		t.Fatal("no error")
	}
}
