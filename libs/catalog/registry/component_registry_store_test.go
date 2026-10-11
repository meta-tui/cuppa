package registry

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func sample() []definition.Definition {
	return []definition.Definition{
		{ID: "bubbles.spinner", Name: "Spinner", Family: definition.FamilyBubbles},
		{ID: "lipgloss.label", Name: "Label", Family: definition.FamilyLipgloss},
		{ID: "lipgloss.box", Name: "Box", Family: definition.FamilyLipgloss, Description: "bordered"},
	}
}

func TestNewRejectsDuplicatesAndEmptyIDs(t *testing.T) {
	if _, err := New(definition.Definition{Name: "x"}); err == nil {
		t.Fatal("empty id accepted")
	}
	d := definition.Definition{ID: "a"}
	if _, err := New(d, d); err == nil {
		t.Fatal("duplicate id accepted")
	}
}

func TestListOrdersByFamilyThenName(t *testing.T) {
	r, err := New(sample()...)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range r.List() {
		got = append(got, d.ID)
	}
	want := []string{"lipgloss.box", "lipgloss.label", "bubbles.spinner"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
	if fs := r.Families(); len(fs) != 2 || fs[0] != definition.FamilyLipgloss {
		t.Fatalf("families = %v", fs)
	}
}

func TestSearchAndGet(t *testing.T) {
	r, _ := New(sample()...)
	if got := r.Search("BORDER"); len(got) != 1 || got[0].ID != "lipgloss.box" {
		t.Fatalf("search = %v", got)
	}
	if got := r.Search(""); len(got) != 3 {
		t.Fatalf("empty search = %d", len(got))
	}
	if _, ok := r.Get("nope"); ok {
		t.Fatal("unknown id found")
	}
}

func TestPacksKeepTheirOrderAndNames(t *testing.T) {
	packs := []definition.Pack{
		{ID: "mine", Name: "My pack"},
		{ID: definition.FamilyBubbles, Name: "Bubbles"},
	}
	r, err := NewWithPacks(packs,
		definition.Definition{ID: "bubbles.spinner", Name: "Spinner", Family: definition.FamilyBubbles},
		definition.Definition{ID: "mine.card", Name: "Card", Family: "mine"},
		definition.Definition{ID: "lipgloss.box", Name: "Box", Family: definition.FamilyLipgloss},
	)
	if err != nil {
		t.Fatal(err)
	}
	fams := r.Families()
	if len(fams) != 3 || fams[0] != "mine" || fams[1] != definition.FamilyBubbles || fams[2] != definition.FamilyLipgloss {
		t.Fatalf("families = %v", fams)
	}
	if r.Title("mine") != "My pack" || r.Title(definition.FamilyLipgloss) != "Lip Gloss" {
		t.Fatalf("titles: %q %q", r.Title("mine"), r.Title(definition.FamilyLipgloss))
	}
	if _, err := NewWithPacks([]definition.Pack{{ID: "a"}, {ID: "a"}}); err == nil {
		t.Fatal("duplicate pack accepted")
	}
}

func TestOnlyHidesDisabledPacksWithoutTouchingTheOriginal(t *testing.T) {
	r, _ := New(sample()...)
	only := r.Only(func(f definition.Family) bool { return f != definition.FamilyBubbles })
	if _, ok := only.Get("bubbles.spinner"); ok {
		t.Fatal("disabled pack's component still listed")
	}
	if got := only.Search(""); len(got) != 2 {
		t.Fatalf("search = %d", len(got))
	}
	if len(only.Families()) != 1 {
		t.Fatalf("families = %v", only.Families())
	}
	if _, ok := r.Get("bubbles.spinner"); !ok {
		t.Fatal("the full registry must still resolve it")
	}
}

func TestLiveFollowsTheRegistryItHolds(t *testing.T) {
	a, _ := New(definition.Definition{ID: "lipgloss.box", Name: "Box", Family: definition.FamilyLipgloss})
	b, _ := New(definition.Definition{ID: "bubbles.spinner", Name: "Spinner", Family: definition.FamilyBubbles})
	live := NewLive(a)
	if _, ok := live.Get("lipgloss.box"); !ok {
		t.Fatal("answers like a")
	}
	live.Set(b)
	if _, ok := live.Get("lipgloss.box"); ok {
		t.Fatal("a is gone")
	}
	if len(live.Families()) != 1 || live.Title(definition.FamilyBubbles) != "Bubbles" || len(live.Packs()) != 1 || len(live.List()) != 1 || len(live.Search("spin")) != 1 || len(live.ByFamily(definition.FamilyBubbles)) != 1 {
		t.Fatal("every query follows the held registry")
	}
	if live.Registry() != b {
		t.Fatal("Registry returns the held one")
	}
}
