package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Suffix is what is added to the name of the design file to name its local
// override: demo.cuppa has demo.cuppa.run.json.
const Suffix = ".run.json"

// PathOf is where the local override of the design at designPath lives.
func PathOf(designPath string) string { return designPath + Suffix }

// Read returns the override file of the design, and an empty one when there is
// none.
func Read(designPath string) (Override, error) {
	data, err := os.ReadFile(PathOf(designPath))
	if errors.Is(err, fs.ErrNotExist) {
		return Override{}, nil
	}
	if err != nil {
		return Override{}, fmt.Errorf("local run settings: %w", err)
	}
	var o Override
	if err := json.Unmarshal(data, &o); err != nil {
		return Override{}, fmt.Errorf("local run settings %s: %w", PathOf(designPath), err)
	}
	return o, nil
}

// Load returns the design's run block with the local override of this
// computer applied. A design with no override file is returned as it is.
func Load(designPath string, base design.RunSettings) (design.RunSettings, error) {
	o, err := Read(designPath)
	if err != nil {
		return base, err
	}
	return Apply(base, o), nil
}

// Save writes the override beside the design.
func Save(designPath string, o Override) error {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(PathOf(designPath), append(data, '\n'), 0o644)
}

// Apply is base with the override's fields on top; env is merged by name.
func Apply(base design.RunSettings, o Override) design.RunSettings {
	if len(o.Build) > 0 {
		base.Build = o.Build
	}
	if len(o.Run) > 0 {
		base.Run = o.Run
	}
	if len(o.Env) > 0 {
		env := make(map[string]string, len(base.Env)+len(o.Env))
		for k, v := range base.Env {
			env[k] = v
		}
		for k, v := range o.Env {
			env[k] = v
		}
		base.Env = env
	}
	if o.Dir != nil {
		base.Dir = *o.Dir
	}
	if o.BuildDir != nil {
		base.BuildDir = *o.BuildDir
	}
	if o.Width != nil {
		base.Width = *o.Width
	}
	if o.Height != nil {
		base.Height = *o.Height
	}
	if o.OnSave != nil {
		base.OnSave = *o.OnSave
	}
	return base.Clean()
}
