package design

import "testing"

func TestRunSettingsNeedAProgramToRun(t *testing.T) {
	if (RunSettings{Build: []string{"go", "build"}}).Clean().IsSet() {
		t.Fatal("a build with nothing to run is a run block")
	}
	s := RunSettings{Run: []string{" ", "x"}}.Clean()
	if s.IsSet() {
		t.Fatalf("a blank program was kept: %+v", s)
	}
	s = RunSettings{Build: []string{"go", "build", "./..."}, Run: []string{"app.exe", "--demo"}, OnSave: true}.Clean()
	if !s.IsSet() || len(s.Build) != 3 || !s.OnSave {
		t.Fatalf("got %+v", s)
	}
}

func TestRunSettingsDropWhatCannotBeUsed(t *testing.T) {
	s := RunSettings{
		Run:    []string{"app"},
		Build:  []string{""},
		Env:    map[string]string{"HOME": "/tmp/x", "": "v", "A=B": "v", "NUL": "a\x00b"},
		Width:  9000,
		Height: -3,
	}.Clean()
	if len(s.Build) != 0 {
		t.Errorf("a build with no program was kept: %v", s.Build)
	}
	if len(s.Env) != 1 || s.Env["HOME"] != "/tmp/x" {
		t.Errorf("env %v", s.Env)
	}
	if s.Width != MaxWidth || s.Height != 0 {
		t.Errorf("size %dx%d", s.Width, s.Height)
	}
}

func TestRunSettingsAreEqualWhenTheyRunTheSameThing(t *testing.T) {
	a := RunSettings{Run: []string{"app"}, Env: map[string]string{"A": "1"}}
	b := RunSettings{Run: []string{"app"}, Env: map[string]string{"A": "1"}}
	if !a.Equal(b) {
		t.Error("equal settings differ")
	}
	b.Env["A"] = "2"
	if a.Equal(b) {
		t.Error("another env is the same")
	}
}
