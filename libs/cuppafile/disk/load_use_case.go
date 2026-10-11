package disk

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/meta-tui/cuppa/libs/cuppafile/format"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// Load reads the design stored at path: a .cuppa file, or the Go source of an
// application Cuppa exported (its main.go, or the file with the design in
// it), which carries the design it was made from (ADR 0012).
func Load(path string) (design.Document, error) {
	if strings.EqualFold(filepath.Ext(path), ".go") {
		return loadFromGo(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return design.Document{}, err
	}
	return format.Decode(data)
}

// IsGoSource reports whether path names a Go file, which Load reads the design
// out of; such a design has no .cuppa file to save over.
func IsGoSource(path string) bool { return strings.EqualFold(filepath.Ext(path), ".go") }

// maxGoSource bounds a Go file read looking for a design.
const maxGoSource = 32 << 20

// loadFromGo reads the design from the file itself or, when the file is a
// main.go that only runs the application, from the Go files beside it: the
// router carries the design.
func loadFromGo(path string) (design.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return design.Document{}, err
	}
	doc, err := format.FromGo(data)
	if !errors.Is(err, format.ErrNoDesign) {
		return doc, err
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return design.Document{}, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".go") || e.Name() == filepath.Base(path) {
			continue
		}
		info, err := e.Info()
		if err != nil || info.Size() > maxGoSource {
			continue
		}
		sibling, err := os.ReadFile(filepath.Join(filepath.Dir(path), e.Name()))
		if err != nil {
			continue
		}
		if doc, err := format.FromGo(sibling); err == nil {
			return doc, nil
		}
	}
	return design.Document{}, format.ErrNoDesign
}
