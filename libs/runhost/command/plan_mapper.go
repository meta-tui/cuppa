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
	dir := designDir
	if s.Dir != "" {
		dir = s.Dir
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(designDir, dir)
		}
	}
	dir = filepath.Clean(dir)
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
		p.Build = &Command{Argv: s.Build, Dir: dir, Env: env}
	}
	return p, nil
}
