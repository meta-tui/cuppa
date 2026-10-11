// Package bundled holds the .cupp packs that ship inside Cuppa. Every file in
// packs/ is embedded at build time and listed with the built-in packs.
package bundled

import (
	"embed"
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
)

//go:embed packs/*.cupp
var files embed.FS

// Packs returns the bundled packs in file name order. A file that cannot be
// read is reported (a test makes that a build failure, so users never see it).
func Packs() ([]cupp.Pack, []error) {
	entries, err := files.ReadDir("packs")
	if err != nil {
		return nil, []error{err}
	}
	var packs []cupp.Pack
	var problems []error
	for _, e := range entries {
		data, err := files.ReadFile("packs/" + e.Name())
		if err != nil {
			problems = append(problems, err)
			continue
		}
		p, issues, err := cupp.Decode(data)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		for _, issue := range issues {
			problems = append(problems, fmt.Errorf("%s: %w", e.Name(), issue))
		}
		packs = append(packs, p)
	}
	return packs, problems
}
