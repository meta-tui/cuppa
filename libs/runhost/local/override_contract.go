// Package local reads the file beside a design that overrides its run block on
// one computer: where the scratch folders are, which build tool is installed.
// The design file is committed and shared; this file is not.
package local

// Override is the content of the local file. A field that is absent leaves
// the design's own value; Env is merged by name.
type Override struct {
	Build    []string          `json:"build,omitempty"`
	Run      []string          `json:"run,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Dir      *string           `json:"dir,omitempty"`
	BuildDir *string           `json:"buildDir,omitempty"`
	Width    *int              `json:"width,omitempty"`
	Height   *int              `json:"height,omitempty"`
	OnSave   *bool             `json:"onSave,omitempty"`
}
