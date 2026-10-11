package gosource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ModUpdate is what go.mod of an exported project lacks or keeps without need.
type ModUpdate struct {
	// Add are requirement lines (module and version) the project needs.
	Add []string
	// Remove are requirement lines of modules only a component that is no
	// longer in the design needed.
	Remove []string
}

// Empty says go.mod is as the project needs it.
func (u ModUpdate) Empty() bool { return len(u.Add) == 0 && len(u.Remove) == 0 }

// PlanMod compares the go.mod in dir with what the project needs. go.mod is
// written once and then belongs to the person, so this only reports; the
// caller asks before ApplyMod. A project that writes no go.mod (a library: the
// module is the host's) or a folder that has none has nothing to plan. Only the
// modules Cuppa itself adds for a component are ever reported for removal.
func PlanMod(dir string, p Project) ModUpdate {
	if _, ok := p.Files["go.mod"]; !ok {
		return ModUpdate{}
	}
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ModUpdate{}
	}
	have := map[string]string{}
	for _, line := range requireLines(string(data)) {
		have[modulePath(line)] = line
	}
	needed := map[string]bool{}
	var u ModUpdate
	for _, r := range p.Requires {
		needed[modulePath(r)] = true
		if _, ok := have[modulePath(r)]; !ok {
			u.Add = append(u.Add, r)
		}
	}
	ours := map[string]bool{}
	for _, e := range extensions {
		if e.require != "" {
			ours[modulePath(e.require)] = true
		}
	}
	for path, line := range have {
		if ours[path] && !needed[path] {
			u.Remove = append(u.Remove, line)
		}
	}
	sort.Strings(u.Add)
	sort.Strings(u.Remove)
	return u
}

// ApplyMod makes the changes of a plan in the go.mod of dir.
func ApplyMod(dir string, u ModUpdate) error {
	path := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	drop := map[string]bool{}
	for _, r := range u.Remove {
		drop[modulePath(r)] = true
	}
	var out []string
	inBlock, inserted := false, len(u.Add) == 0
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "require ("):
			inBlock = true
		case inBlock && trimmed == ")":
			inBlock = false
			if !inserted {
				for _, a := range u.Add {
					out = append(out, "\t"+a)
				}
				inserted = true
			}
		case inBlock && drop[modulePath(trimmed)]:
			continue
		case strings.HasPrefix(trimmed, "require ") && drop[modulePath(strings.TrimPrefix(trimmed, "require "))]:
			continue
		}
		out = append(out, line)
	}
	text := strings.Join(out, "\n")
	if !inserted {
		text = strings.TrimRight(text, "\n") + "\n\nrequire (\n\t" + strings.Join(u.Add, "\n\t") + "\n)\n"
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// Tidy runs go mod tidy in dir, which settles go.mod and writes go.sum. It
// returns what the command printed.
func Tidy(dir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("go mod tidy: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// requireLines are the requirement lines of a go.mod, as "module version".
func requireLines(mod string) []string {
	var out []string
	inBlock := false
	for _, line := range strings.Split(strings.ReplaceAll(mod, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "require ("):
			inBlock = true
		case inBlock && t == ")":
			inBlock = false
		case inBlock && t != "" && !strings.HasPrefix(t, "//"):
			out = append(out, t)
		case strings.HasPrefix(t, "require "):
			out = append(out, strings.TrimSpace(strings.TrimPrefix(t, "require ")))
		}
	}
	return out
}

func modulePath(line string) string {
	if f := strings.Fields(line); len(f) > 0 {
		return f[0]
	}
	return ""
}
