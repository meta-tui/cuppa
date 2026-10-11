package cupp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/registry"
)

// LoadDir reads every .cupp file in dir, in name order. A file that cannot be
// read is skipped and reported; a missing folder is not an error.
func LoadDir(dir string) ([]Pack, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), Extension) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var packs []Pack
	var problems []error
	for _, name := range names {
		p, issues, err := LoadFile(filepath.Join(dir, name))
		if err != nil {
			problems = append(problems, err)
			continue
		}
		problems = append(problems, issues...)
		packs = append(packs, p)
	}
	return packs, problems
}

// UserDir is the folder of installed packs, or "" when the user has no config
// directory.
func UserDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "cuppa", "packs")
}

// LoadUser extends base with the packs installed in UserDir, reporting any
// file that could not be used.
func LoadUser(base *registry.Registry) (*registry.Registry, []error) {
	dir := UserDir()
	if dir == "" {
		return base, nil
	}
	packs, problems := LoadDir(dir)
	reg, more := Extend(base, packs)
	return reg, append(problems, more...)
}

// LoadFile reads one .cupp file.
func LoadFile(path string) (Pack, []error, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Pack{}, nil, err
	}
	p, issues, err := Decode(data)
	if err != nil {
		return Pack{}, nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	p.Path = path
	for i, e := range issues {
		issues[i] = fmt.Errorf("%s: %w", filepath.Base(path), e)
	}
	return p, issues, nil
}

// Extend returns base with the packs added. A pack whose id is already taken
// (by a built-in or an earlier pack) is left out and reported.
func Extend(base *registry.Registry, packs []Pack) (*registry.Registry, []error) {
	all := base.Packs()
	defs := base.List()
	taken := map[string]bool{}
	for _, p := range all {
		taken[string(p.ID)] = true
	}
	var problems []error
	for _, p := range packs {
		if taken[p.ID] {
			problems = append(problems, fmt.Errorf("pack %q: id already in use", p.ID))
			continue
		}
		taken[p.ID] = true
		pack, d := Definitions(p)
		all = append(all, pack)
		defs = append(defs, d...)
	}
	r, err := registry.NewWithPacks(all, defs...)
	if err != nil {
		return base, append(problems, err)
	}
	return r, problems
}
