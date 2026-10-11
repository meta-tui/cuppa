package gosource

import (
	"os"
	"path/filepath"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// ExportDir is the folder the settings name, resolved against the folder of the
// design file (an empty path resolves against the working directory).
func ExportDir(s design.ExportSettings, designPath string) string {
	if filepath.IsAbs(s.Dir) || designPath == "" {
		return filepath.Clean(s.Dir)
	}
	return filepath.Join(filepath.Dir(designPath), s.Dir)
}

// GenerateExport generates the project the settings ask for. source is the
// design as the bytes of a .cuppa file, which an application carries so it can
// be opened again.
func GenerateExport(doc design.Document, cat Catalog, s design.ExportSettings, source []byte) (Project, error) {
	opts := Options{Source: source, Module: s.Module, Library: s.Library}
	if s.Layout == design.ExportSlices {
		return GenerateSlices(doc, cat, opts)
	}
	return GenerateWith(doc, cat, opts), nil
}

// WriteExport writes a project made by GenerateExport. Exporting again
// replaces what Cuppa wrote earlier, so the design can be edited and exported
// as often as needed. A slice export keeps the person's own files and removes
// the generated ones that are gone; the one-folder layout, whose files carry no
// header, rewrites the files it has and removes nothing. go.mod and README.md
// are written only the first time.
func WriteExport(dir string, s design.ExportSettings, p Project) error {
	return WriteExportKeeping(dir, s, p, nil)
}

// ObsoleteExport are the generated files WriteExport would remove from dir: the
// ones of a page or a component the design no longer has. The one-folder layout
// removes nothing.
func ObsoleteExport(dir string, s design.ExportSettings, p Project) []string {
	if s.Layout != design.ExportSlices {
		return nil
	}
	return Obsolete(dir, p)
}

// WriteExportKeeping is WriteExport that leaves the files named in keep, which
// ObsoleteExport reported, in place.
func WriteExportKeeping(dir string, s design.ExportSettings, p Project, keep []string) error {
	if s.Layout == design.ExportSlices {
		return ReplaceProjectKeeping(dir, p, keep)
	}
	for name, content := range p.Files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if _, err := os.Stat(path); err == nil && keepsOwn[name] {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
