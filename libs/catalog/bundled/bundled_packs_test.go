package bundled

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

// A pack added to packs/ must be clean and must not clash with anything else
// that ships, so a bad contribution fails here, in CI, before it can be merged.
func TestEveryBundledPackIsCleanAndUnique(t *testing.T) {
	packs, problems := Packs()
	for _, p := range problems {
		t.Errorf("bundled pack problem: %v", p)
	}
	if len(packs) == 0 {
		t.Fatal("there should be at least the starter pack")
	}
	seen := map[string]bool{}
	for _, p := range standard.Packs() {
		seen[string(p.ID)] = true
	}
	for _, p := range packs {
		if seen[p.ID] {
			t.Errorf("pack id %q is used twice", p.ID)
		}
		seen[p.ID] = true
		if p.Description == "" || p.Version == "" {
			t.Errorf("pack %q needs a description and a version", p.ID)
		}
		if len(p.Components) == 0 {
			t.Errorf("pack %q has no components", p.ID)
		}
	}
	reg, issues := cupp.Extend(standard.Default(), packs)
	if len(issues) != 0 {
		t.Errorf("the packs do not assemble: %v", issues)
	}
	// Parts may only use components that ship: built-ins or bundled packs.
	for _, d := range reg.List() {
		if d.Inner == nil {
			continue
		}
		for _, part := range d.Inner.Nodes {
			if _, ok := reg.Get(part.Component); !ok {
				t.Errorf("%s uses %q, which does not ship with Cuppa", d.ID, part.Component)
			}
		}
	}
}
