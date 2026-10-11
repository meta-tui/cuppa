package design

import "strings"

// Layouts an export can have (ADR 0013).
const (
	ExportFlat   = "flat"
	ExportSlices = "slices"
)

// ExportSettings say where the design's Go source goes, so the design is the
// one file that knows the structure of the application it draws: the person
// opens it, edits the layout and exports again into the same folders.
type ExportSettings struct {
	// Layout is ExportFlat or ExportSlices.
	Layout string `json:"layout,omitempty"`
	// Library writes the slices into a module that exists (no go.mod, main.go
	// or router). It needs ExportSlices.
	Library bool `json:"library,omitempty"`
	// Dir is the folder the files go to, relative to the folder of the design
	// file (or absolute).
	Dir string `json:"dir,omitempty"`
	// Module is the import path of Dir (slices), or the module's name.
	Module string `json:"module,omitempty"`
}

// IsSet reports whether the design says where to export.
func (s ExportSettings) IsSet() bool { return s.Layout != "" && s.Dir != "" }

// Clean returns the settings with what cannot be used taken away.
func (s ExportSettings) Clean() ExportSettings {
	s.Dir = strings.TrimSpace(s.Dir)
	s.Module = strings.TrimSpace(s.Module)
	if s.Layout != ExportFlat && s.Layout != ExportSlices || len(s.Dir) > 400 || strings.ContainsRune(s.Dir, 0) {
		return ExportSettings{}
	}
	if len(s.Module) > 200 {
		s.Module = ""
	}
	if s.Layout != ExportSlices {
		s.Library = false
	}
	return s
}
