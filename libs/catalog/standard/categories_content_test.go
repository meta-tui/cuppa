package standard

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func TestEveryListedComponentHasACategoryOtherThanOther(t *testing.T) {
	for _, d := range All() {
		if d.Hidden {
			continue
		}
		if d.Category == "" || d.Category == definition.CategoryOther {
			t.Errorf("%s has no category: add it to categories", d.ID)
		}
	}
}

func TestEveryCategoryNamesAComponentThatExists(t *testing.T) {
	r := Default()
	for id := range categories {
		if _, ok := r.Get(id); !ok {
			t.Errorf("categories names %q, which is not a component", id)
		}
	}
}

func TestTheSameKindFromDifferentProvidersSitsTogether(t *testing.T) {
	r := Default()
	var from = map[definition.Family]bool{}
	for _, d := range r.ByCategory(definition.CategoryData) {
		if d.Name == "List" {
			from[d.Family] = true
		}
	}
	if !from[definition.FamilyBubbles] || !from[definition.FamilyLipgloss] {
		t.Errorf("Data should hold the Bubbles and the Lip Gloss list, got %v", from)
	}
}
