// Package packstate remembers which component packs the user switched off.
package packstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

// State is the set of disabled packs, optionally kept in a file. Packs are on
// unless switched off, so a new pack appears without any bookkeeping.
type State struct {
	path     string
	disabled map[definition.Family]bool
}

type fileBody struct {
	Disabled []string `json:"disabled"`
}

// Open reads the state kept at path. A missing or damaged file means every
// pack is on. An empty path keeps the state in memory only.
func Open(path string) *State {
	s := &State{path: path, disabled: map[definition.Family]bool{}}
	if path == "" {
		return s
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	var body fileBody
	if json.Unmarshal(data, &body) != nil {
		return s
	}
	for _, id := range body.Disabled {
		s.disabled[definition.Family(id)] = true
	}
	return s
}

// Enabled reports whether the pack's components appear in the palette.
func (s *State) Enabled(id definition.Family) bool { return !s.disabled[id] }

// SetEnabled switches a pack on or off and saves the choice.
func (s *State) SetEnabled(id definition.Family, on bool) {
	if on == s.Enabled(id) {
		return
	}
	if on {
		delete(s.disabled, id)
	} else {
		s.disabled[id] = true
	}
	s.save()
}

func (s *State) save() {
	if s.path == "" {
		return
	}
	body := fileBody{Disabled: []string{}}
	for id := range s.disabled {
		body.Disabled = append(body.Disabled, string(id))
	}
	sort.Strings(body.Disabled)
	data, err := json.Marshal(body)
	if err != nil || os.MkdirAll(filepath.Dir(s.path), 0o755) != nil {
		return
	}
	_ = os.WriteFile(s.path, data, 0o644)
}
