package command

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// PlanOf is the plan for a run block of the design saved in designDir. A
// relative working directory is relative to designDir. It fails when the block
// says nothing to run.
func PlanOf(s design.RunSettings, designDir string) (Plan, error) {
	s = s.Clean()
	if !s.IsSet() {
		return Plan{}, fmt.Errorf("runhost: the design does not say what to run")
	}
	dir := resolveDir(s.Dir, designDir)
	buildDir := dir
	if s.BuildDir != "" {
		buildDir = resolveDir(s.BuildDir, designDir)
	}
	env := make([]string, 0, len(s.Env))
	for name, value := range s.Env {
		env = append(env, name+"="+value)
	}
	sort.Strings(env)
	p := Plan{
		Run:    Command{Argv: s.Run, Dir: dir, Env: env},
		Width:  s.Width,
		Height: s.Height,
		OnSave: s.OnSave,
	}
	if len(s.Build) > 0 {
		p.Build = &Command{Argv: s.Build, Dir: buildDir, Env: env}
	}
	return p, nil
}

// resolveDir is dir made absolute against the folder of the design; empty is
// that folder.
func resolveDir(dir, designDir string) string {
	if dir == "" {
		return filepath.Clean(designDir)
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(designDir, dir)
	}
	return filepath.Clean(dir)
}
