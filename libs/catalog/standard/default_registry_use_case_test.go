package standard

import (
	"strconv"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func TestShippedDefinitionsAreConsistent(t *testing.T) {
	r := Default()
	if len(r.List()) == 0 {
		t.Fatal("empty catalog")
	}
	for _, d := range r.List() {
		if d.Name == "" || d.Import == "" || d.Status == "" {
			t.Errorf("%s: missing name, import or status", d.ID)
		}
		if d.DefaultSize.W < d.MinSize.W || d.DefaultSize.H < d.MinSize.H || d.MinSize.W < 1 || d.MinSize.H < 1 {
			t.Errorf("%s: bad sizes %+v / %+v", d.ID, d.DefaultSize, d.MinSize)
		}
		seen := map[string]bool{}
		for _, p := range d.Props {
			if seen[p.Key] {
				t.Errorf("%s: duplicate prop %s", d.ID, p.Key)
			}
			seen[p.Key] = true
			switch p.Kind {
			case definition.PropChoice:
				if !contains(p.Choices, p.Default) {
					t.Errorf("%s.%s: default %q not in choices", d.ID, p.Key, p.Default)
				}
			case definition.PropInt:
				n, err := strconv.Atoi(p.Default)
				if err != nil || (p.Max > p.Min && (n < p.Min || n > p.Max)) {
					t.Errorf("%s.%s: bad int default %q", d.ID, p.Key, p.Default)
				}
			case definition.PropBool:
				if p.Default != "true" && p.Default != "false" {
					t.Errorf("%s.%s: bad bool default %q", d.ID, p.Key, p.Default)
				}
			}
		}
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
