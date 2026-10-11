package cupp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func samplePack() Pack {
	return Pack{
		ID: "tea-shop", Name: "Tea shop", Version: "1.0.0", Description: "Cards for a tea shop",
		Components: []design.Composite{{
			ID: "card", Name: "Card", W: 20, H: 5,
			Nodes: []design.Node{
				{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}},
				{ID: "b", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 2, Y: 1, W: 10, H: 1}},
			},
			Props: []design.Exposed{{Key: "title", Label: "Title", Kind: "text", Default: "Tea", Target: "b", TargetProp: "text"}},
		}},
	}
}

func TestPackRoundTrips(t *testing.T) {
	data, err := Encode(samplePack())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte(Magic)) {
		t.Fatal("the file starts with the magic")
	}
	got, problems, err := Decode(data)
	if err != nil || len(problems) != 0 {
		t.Fatalf("decode: %v %v", err, problems)
	}
	if !reflect.DeepEqual(got, samplePack()) {
		t.Fatalf("round trip changed the pack:\n%+v", got)
	}
}

func TestEncodeRefusesAPackItCouldNotReadBack(t *testing.T) {
	p := samplePack()
	p.ID = "Bad ID"
	if _, err := Encode(p); err == nil {
		t.Fatal("bad pack id")
	}
	p = samplePack()
	p.Components = append(p.Components, p.Components[0])
	if _, err := Encode(p); err == nil {
		t.Fatal("repeated component id")
	}
}

func TestDecodeRejectsWhatIsNotAPack(t *testing.T) {
	good, _ := Encode(samplePack())
	if _, _, err := Decode([]byte("CUPPA\n\x00\x01")); !errors.Is(err, ErrNotCupp) {
		t.Errorf("a .cuppa file is not a pack: %v", err)
	}
	if _, _, err := Decode(append([]byte(Magic), 0)); !errors.Is(err, ErrCorrupt) {
		t.Errorf("short header: %v", err)
	}
	newer := append([]byte(nil), good...)
	newer[len(Magic)+1] = 9
	if _, _, err := Decode(newer); !errors.Is(err, ErrTooNew) {
		t.Errorf("newer version: %v", err)
	}
	zero := append([]byte(nil), good...)
	zero[len(Magic)], zero[len(Magic)+1] = 0, 0
	if _, _, err := Decode(zero); !errors.Is(err, ErrCorrupt) {
		t.Errorf("version zero: %v", err)
	}
	if _, _, err := Decode(good[:len(good)-6]); !errors.Is(err, ErrCorrupt) {
		t.Errorf("truncated: %v", err)
	}
}

// rawFile builds a file without Encode's checks, the way a hand edit would.
func rawFile(t *testing.T, p Pack) []byte {
	t.Helper()
	body, err := json.Marshal(envelope{Pack: p})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	out.WriteString(Magic)
	out.Write([]byte{0, 1})
	zw := gzip.NewWriter(&out)
	if _, err := zw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestOneBadComponentDoesNotCostThePack(t *testing.T) {
	p := samplePack()
	broken := p.Components[0].Clone()
	broken.ID = "broken"
	broken.Props[0].Target = "nowhere"
	twin := p.Components[0].Clone() // the same id as the first
	p.Components = append(p.Components, broken, twin)
	got, problems, err := Decode(rawFile(t, p))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Components) != 1 || got.Components[0].ID != "card" || len(problems) != 2 {
		t.Fatalf("kept %d components, problems %v", len(got.Components), problems)
	}
	if _, err := Encode(p); err == nil {
		t.Fatal("Encode must refuse the same pack")
	}
}

func FuzzDecode(f *testing.F) {
	good, _ := Encode(samplePack())
	f.Add(good)
	f.Add([]byte(Magic))
	f.Add([]byte("junk"))
	f.Fuzz(func(t *testing.T, data []byte) {
		p, _, err := Decode(data)
		if err == nil && !design.ValidID(p.ID) {
			t.Fatalf("a decoded pack always has a valid id: %q", p.ID)
		}
	})
}

func TestLoadDirReadsPacksAndReportsBadFiles(t *testing.T) {
	dir := t.TempDir()
	data, _ := Encode(samplePack())
	if err := os.WriteFile(filepath.Join(dir, "tea.cupp"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.cupp"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}
	packs, problems := LoadDir(dir)
	if len(packs) != 1 || packs[0].ID != "tea-shop" {
		t.Fatalf("packs = %+v", packs)
	}
	if len(problems) != 1 {
		t.Fatalf("the bad file is reported: %v", problems)
	}
	if packs, problems := LoadDir(filepath.Join(dir, "missing")); packs != nil || problems != nil {
		t.Fatal("a missing folder is not an error")
	}
}

func TestExtendAddsPacksAndRefusesIDClashes(t *testing.T) {
	base := standard.Default()
	clash := samplePack()
	clash.ID = "bubbles"
	reg, problems := Extend(base, []Pack{samplePack(), clash})
	if len(problems) != 1 {
		t.Fatalf("the clash is reported: %v", problems)
	}
	def, ok := reg.Get("tea-shop.card")
	if !ok || def.Inner == nil || def.DefaultSize.W != 20 || len(def.Props) != 1 {
		t.Fatalf("definition = %+v", def)
	}
	if reg.Title("tea-shop") != "Tea shop" {
		t.Fatal("the pack is listed under its name")
	}
	if _, ok := reg.Get("bubbles.spinner"); !ok {
		t.Fatal("built-ins stay")
	}
	if _, ok := base.Get("tea-shop.card"); ok {
		t.Fatal("the base registry is untouched")
	}
}

func TestLoadedPacksRememberTheirFile(t *testing.T) {
	dir := t.TempDir()
	data, _ := Encode(samplePack())
	path := filepath.Join(dir, "tea.cupp")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	packs, _ := LoadDir(dir)
	reg, _ := Extend(standard.Default(), packs)
	for _, p := range reg.Packs() {
		if p.ID == "tea-shop" && (p.Source != path || p.Builtin) {
			t.Fatalf("pack = %+v", p)
		}
	}
}

func TestJSONFormRoundTripsAndNamesTheFirstProblem(t *testing.T) {
	data, err := ToJSON(samplePack())
	if err != nil {
		t.Fatal(err)
	}
	got, err := FromJSON(data)
	if err != nil || !reflect.DeepEqual(got, samplePack()) {
		t.Fatalf("round trip: %v\n%+v", err, got)
	}
	if _, err := FromJSON([]byte("{")); err == nil {
		t.Fatal("broken JSON")
	}
	bad := samplePack()
	bad.Components[0].Props[0].Target = "nowhere"
	data, _ = ToJSON(bad)
	if _, err := FromJSON(data); err == nil {
		t.Fatal("a pack with a bad component is refused, not trimmed")
	}
}
