package registry

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// Live is a catalog that can be swapped for a new registry while the app runs
// (a pack was added, removed or created). Everything that holds the Live sees
// the change; it answers like the registry it currently holds.
type Live struct{ r *Registry }

// NewLive returns a Live holding r.
func NewLive(r *Registry) *Live { return &Live{r: r} }

// Set replaces the registry.
func (l *Live) Set(r *Registry) { l.r = r }

// Registry returns the registry currently held.
func (l *Live) Registry() *Registry { return l.r }

// Get returns the definition with the given id.
func (l *Live) Get(id string) (definition.Definition, bool) { return l.r.Get(id) }

// List returns every definition.
func (l *Live) List() []definition.Definition { return l.r.List() }

// Families returns the packs that have components.
func (l *Live) Families() []definition.Family { return l.r.Families() }

// ByFamily returns the definitions of one pack.
func (l *Live) ByFamily(f definition.Family) []definition.Definition { return l.r.ByFamily(f) }

// Search returns the definitions matching query.
func (l *Live) Search(query string) []definition.Definition { return l.r.Search(query) }

// Title is the display name of a pack.
func (l *Live) Title(f definition.Family) string { return l.r.Title(f) }

// Packs returns every pack.
func (l *Live) Packs() []definition.Pack { return l.r.Packs() }
