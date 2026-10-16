// Package component reads the Cuppa component description: the
// cuppa.component.json file a Go module publishes next to its code so that a
// design tool can list the component, show its properties and generate the Go
// that uses it. See docs/spec/cuppa-component.md.
package component

// FileName is the name of the description file at the root of a repository.
const FileName = "cuppa.component.json"

// Version is the version of the description format this package reads.
const Version = 1

// Description is one component, as its module describes it.
type Description struct {
	// Cuppa is the version of the format (1).
	Cuppa int `json:"cuppa"`
	// Name is the component's short name: lowercase letters, digits and dashes.
	// It is unique within the module's catalog family.
	Name string `json:"name"`
	// Title is what the palette shows.
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	License     string `json:"license"`
	Homepage    string `json:"homepage"`
	// Family groups the component in the palette (default "community").
	Family string `json:"family"`
	// Category says what the component is for: the palette groups by it
	// (default "other").
	Category string `json:"category"`

	Go      GoBinding `json:"go"`
	Size    Size      `json:"size"`
	Props   []Prop    `json:"props"`
	Events  []Event   `json:"events"`
	Preview Preview   `json:"preview"`
}

// GoBinding says how to use the component from Go.
type GoBinding struct {
	// Module is the module path, Import the package's import path.
	Module  string `json:"module"`
	Import  string `json:"import"`
	Package string `json:"package"`
	// BubbleTea is the major version of Bubble Tea the component is for.
	BubbleTea int `json:"bubbletea"`
	// Model is the component's type, Constructor the function that makes one.
	Model       string `json:"model"`
	Constructor string `json:"constructor"`
	// Init, Update, View, Width, Height and Origin name the methods that
	// follow the Bubble Tea shape; empty means the component has none.
	Init   string `json:"init"`
	Update string `json:"update"`
	View   string `json:"view"`
	Width  string `json:"width"`
	Height string `json:"height"`
	Origin string `json:"origin"`
}

// Size is the component's default and smallest size in cells.
type Size struct {
	Default Cells  `json:"default"`
	Min     Cells  `json:"min"`
	Note    string `json:"note"`
}

// Cells is a width and a height.
type Cells struct {
	W int `json:"w"`
	H int `json:"h"`
}

// Prop is a property of the component: what the designer can set.
type Prop struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Kind is text, int, bool, color or choice.
	Kind    string   `json:"kind"`
	Default string   `json:"default"`
	Choices []string `json:"choices"`
	Min     int      `json:"min"`
	Max     int      `json:"max"`
	Go      PropGo   `json:"go"`
}

// PropGo says how a property reaches the Go component.
type PropGo struct {
	// Option is the constructor option that sets it, Setter the method that
	// sets it later and Getter the method that reads it.
	Option string `json:"option"`
	Setter string `json:"setter"`
	Getter string `json:"getter"`
	// Variadic is true when the option takes the list as separate arguments,
	// Split the separator of a list written in one text property.
	Variadic bool   `json:"variadic"`
	Split    string `json:"split"`
}

// Event is a message the component sends.
type Event struct {
	Name   string       `json:"name"`
	Label  string       `json:"label"`
	Go     EventGo      `json:"go"`
	Fields []EventField `json:"fields"`
}

// EventGo names the message type.
type EventGo struct {
	Message string `json:"message"`
}

// EventField is a field of an event's message; Prop names the property it
// carries, if any.
type EventField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Prop string `json:"prop"`
}

// Preview says how a design draws the component. "frame" (the default) is a
// labelled frame; "static" draws Lines.
type Preview struct {
	Kind  string   `json:"kind"`
	Lines []string `json:"lines"`
}
