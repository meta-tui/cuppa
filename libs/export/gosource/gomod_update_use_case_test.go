package gosource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modWithFiglet = "module example.com/app\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbletea/v2 v2.1.0\n\t" + figletRequire + "\n)\n"

func projectNeeding(requires ...string) Project {
	return Project{Files: map[string]string{"go.mod": ""}, Requires: requires}
}

// A requirement the design now needs is added; one only a removed component
// needed is removed; the person's own requirements stay.
func TestGoModIsPlannedAndUpdatedOnlyForWhatCuppaAdds(t *testing.T) {
	dir := t.TempDir()
	mod := strings.Replace(modWithFiglet, "\n)\n", "\n\texample.com/mine v1.0.0\n)\n", 1)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	var other string
	for _, e := range extensions {
		if e.require != "" && e.require != figletRequire {
			other = e.require
			break
		}
	}
	if other == "" {
		t.Skip("no other module in the catalog")
	}
	u := PlanMod(dir, projectNeeding(other))
	if len(u.Add) != 1 || u.Add[0] != other || len(u.Remove) != 1 || u.Remove[0] != figletRequire {
		t.Fatalf("plan: %+v", u)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "go.mod")); string(data) != mod {
		t.Fatal("planning must not write")
	}
	if err := ApplyMod(dir, u); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	got := string(data)
	if !strings.Contains(got, other) || strings.Contains(got, figletRequire) || !strings.Contains(got, "example.com/mine v1.0.0") || !strings.Contains(got, "bubbletea/v2") {
		t.Fatalf("go.mod:\n%s", got)
	}
	if again := PlanMod(dir, projectNeeding(other)); !again.Empty() {
		t.Fatalf("after applying: %+v", again)
	}
}

func TestAProjectWithoutItsOwnGoModHasNothingToPlan(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modWithFiglet), 0o644); err != nil {
		t.Fatal(err)
	}
	if u := PlanMod(dir, Project{Files: map[string]string{}}); !u.Empty() {
		t.Fatalf("library: %+v", u)
	}
	if u := PlanMod(t.TempDir(), projectNeeding(figletRequire)); !u.Empty() {
		t.Fatalf("no go.mod: %+v", u)
	}
}
