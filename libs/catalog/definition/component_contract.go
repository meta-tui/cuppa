// Package definition holds the contracts that describe a designable component.
package definition

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Family groups components by where they come from.
type Family string

// Component families, in the order the palette lists them.
const (
	FamilyLipgloss  Family = "lipgloss"
	FamilyBubbles   Family = "bubbles"
	FamilyHuh       Family = "huh"
	FamilyGlamour   Family = "glamour"
	FamilyNtcharts  Family = "ntcharts"
	FamilyCommunity Family = "community"
)

// FamilyOrder is the display order of the families.
var FamilyOrder = []Family{FamilyLipgloss, FamilyBubbles, FamilyHuh, FamilyGlamour, FamilyNtcharts, FamilyCommunity}

// Title is the human-readable family name.
func (f Family) Title() string {
	switch f {
	case FamilyLipgloss:
		return "Lip Gloss"
	case FamilyBubbles:
		return "Bubbles"
	case FamilyHuh:
		return "Huh forms"
	case FamilyGlamour:
		return "Glamour"
	case FamilyNtcharts:
		return "ntcharts"
	case FamilyCommunity:
		return "Community"
	}
	return string(f)
}

// Status says how faithfully the designer can show a component.
type Status string

// Component statuses.
const (
	// StatusSupported components have a faithful preview.
	StatusSupported Status = "supported"
	// StatusPlaceholder components are drawn as a labelled approximation.
	StatusPlaceholder Status = "placeholder"
	// StatusPlanned components are known but not yet designable.
	StatusPlanned Status = "planned"
)

// PropKind decides how a property is edited and parsed.
type PropKind string

// Property kinds.
const (
	PropText   PropKind = "text"
	PropInt    PropKind = "int"
	PropFloat  PropKind = "float"
	PropBool   PropKind = "bool"
	PropColor  PropKind = "color"
	PropChoice PropKind = "choice"
)

// PropSpec describes one editable property of a component. Values are stored
// as strings and interpreted according to Kind.
type PropSpec struct {
	Key     string
	Label   string
	Kind    PropKind
	Default string
	// Choices lists the allowed values of a PropChoice property.
	Choices []string
	// Min and Max bound a PropInt property when Max > Min.
	Min, Max int
	// Role ties a colour property to a part of the design's theme: until the
	// component sets it, it follows that colour. A value stored on the
	// component, even an empty one, overrides the theme.
	Role Role
}

// MutedKey is the reserved property key under which Effective hands the
// theme's muted colour to painters ("" when the theme sets none). It is not a
// property of any component and is never stored.
const MutedKey = "theme.muted"

// PaletteKey is the reserved property key under which Effective hands the
// design's named colours to painters, as "Name=#hex;Name=#hex". A text
// property that names a colour (@Name) is read through it.
const PaletteKey = "theme.palette"

// Role is the part of a theme a colour property follows when the component
// does not set it itself.
type Role string

// Theme roles. The canvas colour is a role too, for a colour that should be
// the design's own background.
const (
	RoleNone       Role = ""
	RoleText       Role = "text"
	RoleBorder     Role = "border"
	RoleSecondary  Role = "secondary"
	RoleBackground Role = "background"
)

// Size is a width and height in cells.
type Size struct {
	W, H int
}

// Definition describes one component the designer can place.
type Definition struct {
	// ID is stable and unique, for example "bubbles.spinner".
	ID          string
	Name        string
	Family      Family
	Description string
	DefaultSize Size
	MinSize     Size
	Props       []PropSpec
	// Import is the Go import path the real component lives at.
	Import string
	// Hidden components are not listed in the palette or found by search; they
	// are made by a tool (the drawing layer).
	Hidden bool
	Status Status
	// Inner is set on a component made of other components (from a .cupp
	// pack): it is drawn by painting these parts, not by a dedicated painter.
	Inner *design.Composite
}

// Defaults returns the default value of every property.
func (d Definition) Defaults() map[string]string {
	out := make(map[string]string, len(d.Props))
	for _, p := range d.Props {
		out[p.Key] = p.Default
	}
	return out
}

// Effective returns the value of every property of a component placed with
// the values own: its own where it has one, else (for a colour tied to a theme
// role) the theme's colour when the theme sets it, else the default.
func (d Definition) Effective(own map[string]string, theme design.Theme, background string) map[string]string {
	out := make(map[string]string, len(d.Props)+1)
	out[MutedKey] = theme.Muted
	if len(theme.Palette) > 0 {
		entries := make([]string, len(theme.Palette))
		for i, s := range theme.Palette {
			entries[i] = s.Name + "=" + s.Color
		}
		out[PaletteKey] = strings.Join(entries, ";")
	}
	for _, p := range d.Props {
		if v, set := own[p.Key]; set {
			out[p.Key] = v
			continue
		}
		out[p.Key] = p.Default
		var themed string
		switch p.Role {
		case RoleText:
			themed = theme.Text
		case RoleBorder:
			themed = theme.Border
		case RoleSecondary:
			themed = theme.Secondary
		case RoleBackground:
			themed = background
		}
		if themed != "" {
			out[p.Key] = themed
		}
	}
	for k, v := range own { // values for properties the definition does not list
		if _, listed := out[k]; !listed {
			out[k] = v
		}
	}
	// A colour property set to "@Name" uses the palette swatch of that name.
	for _, p := range d.Props {
		if p.Kind == PropColor {
			out[p.Key] = theme.Resolve(out[p.Key])
		}
	}
	return out
}

// Prop returns the specification of the property with the given key.
func (d Definition) Prop(key string) (PropSpec, bool) {
	for _, p := range d.Props {
		if p.Key == key {
			return p, true
		}
	}
	return PropSpec{}, false
}
