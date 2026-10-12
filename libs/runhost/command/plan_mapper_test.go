package command

import (
	"path/filepath"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestARelativeDirectoryIsRelativeToTheDesign(t *testing.T) {
	base := t.TempDir()
	p, err := PlanOf(design.RunSettings{
		Build: []string{"go", "build"},
		Run:   []string{"app"},
		Dir:   "..",
		Env:   map[string]string{"B": "2", "A": "1"},
	}, filepath.Join(base, "designs"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Run.Dir != base || p.Build == nil || p.Build.Dir != base {
		t.Fatalf("dir %q", p.Run.Dir)
	}
	if len(p.Run.Env) != 2 || p.Run.Env[0] != "A=1" || p.Run.Env[1] != "B=2" {
		t.Fatalf("env %v", p.Run.Env)
	}
}

func TestTheBuildAndTheProgramCanRunInDifferentFolders(t *testing.T) {
	base := t.TempDir()
	scratch := filepath.Join(base, "scratch")
	p, err := PlanOf(design.RunSettings{
		Build:    []string{"go", "build"},
		Run:      []string{"app"},
		BuildDir: "..",
		Dir:      scratch,
	}, filepath.Join(base, "designs"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Build.Dir != base || p.Run.Dir != scratch {
		t.Fatalf("build in %q, run in %q", p.Build.Dir, p.Run.Dir)
	}
}

func TestABlockWithNothingToRunHasNoPlan(t *testing.T) {
	if _, err := PlanOf(design.RunSettings{Build: []string{"go", "build"}}, t.TempDir()); err == nil {
		t.Fatal("a plan with nothing to run")
	}
}

func TestNoBuildMeansNothingToBuild(t *testing.T) {
	p, err := PlanOf(design.RunSettings{Run: []string{"app"}}, t.TempDir())
	if err != nil || p.Build != nil {
		t.Fatalf("%+v %v", p, err)
	}
}
