package design

import (
	"slices"
	"strings"
)

// Limits of a run block, so a hand-made file cannot make Cuppa start something
// unbounded.
const (
	maxRunArgs    = 64
	maxRunArgLen  = 1000
	maxRunEnv     = 64
	maxRunDirLen  = 400
	maxRunEnvSize = 2000
)

// RunSettings say how to run the application the design draws, so Cuppa can
// host it in a terminal pane next to the editor and rebuild it when the design
// is saved. Nothing about it is written in the application.
type RunSettings struct {
	// Build is the command that builds the program, as the program and its
	// arguments, run in Dir before Run. Empty when there is nothing to build.
	Build []string `json:"build,omitempty"`
	// Run is the program to host, as the program and its arguments.
	Run []string `json:"run,omitempty"`
	// Env is added to the environment of Build and Run: name to value. It is how
	// the program is pointed at a scratch folder instead of the person's own.
	Env map[string]string `json:"env,omitempty"`
	// Dir is the working directory of Run, relative to the folder of the design
	// file (or absolute). Empty is the folder of the design.
	Dir string `json:"dir,omitempty"`
	// BuildDir is the working directory of Build, resolved the same way. Empty
	// follows Dir.
	BuildDir string `json:"buildDir,omitempty"`
	// Width and Height fix the size of the program's terminal in cells; zero
	// follows the size of the pane.
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	// OnSave builds and relaunches the program when the design is saved.
	OnSave bool `json:"onSave,omitempty"`
}

// IsSet reports whether the design says what to run.
func (s RunSettings) IsSet() bool { return len(s.Run) > 0 }

// Clean returns the settings with what cannot be used taken away.
func (s RunSettings) Clean() RunSettings {
	s.Build = cleanCommand(s.Build)
	s.Run = cleanCommand(s.Run)
	if len(s.Run) == 0 {
		return RunSettings{}
	}
	s.Dir = cleanRunDir(s.Dir)
	s.BuildDir = cleanRunDir(s.BuildDir)
	env := make(map[string]string, len(s.Env))
	for name, value := range s.Env {
		name = strings.TrimSpace(name)
		if name == "" || strings.ContainsAny(name, "=\x00") || len(name) > maxRunArgLen ||
			len(value) > maxRunEnvSize || strings.ContainsRune(value, 0) || len(env) >= maxRunEnv {
			continue
		}
		env[name] = value
	}
	if len(env) == 0 {
		env = nil
	}
	s.Env = env
	s.Width = min(max(s.Width, 0), MaxWidth)
	s.Height = min(max(s.Height, 0), MaxHeight)
	return s
}

// Equal reports whether two settings run the same thing the same way.
func (s RunSettings) Equal(o RunSettings) bool {
	return slices.Equal(s.Build, o.Build) && slices.Equal(s.Run, o.Run) && s.Dir == o.Dir && s.BuildDir == o.BuildDir &&
		s.Width == o.Width && s.Height == o.Height && s.OnSave == o.OnSave && mapsEqual(s.Env, o.Env)
}

func cleanRunDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if len(dir) > maxRunDirLen || strings.ContainsRune(dir, 0) {
		return ""
	}
	return dir
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

// cleanCommand drops a command that has no program, and bounds the rest.
func cleanCommand(argv []string) []string {
	if len(argv) == 0 || strings.TrimSpace(argv[0]) == "" || len(argv) > maxRunArgs {
		return nil
	}
	out := make([]string, 0, len(argv))
	for _, a := range argv {
		if len(a) > maxRunArgLen || strings.ContainsRune(a, 0) {
			return nil
		}
		out = append(out, a)
	}
	return out
}
