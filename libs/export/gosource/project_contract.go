// Package gosource exports a design as a Go program: a Bubble Tea v2 project
// that places the real Bubbles models (and Lip Gloss boxes and labels) where
// the design has them. Components that cannot be generated yet are drawn as an
// empty frame and listed in the notes.
package gosource

// Versions of the libraries the generated project asks for. They are the ones
// Cuppa itself is built with.
const (
	bubbleteaVersion = "v2.1.0"
	lipglossVersion  = "v2.0.6"
	bubblesVersion   = "v2.2.1"
)

// Project is a generated program, ready to write to a folder.
type Project struct {
	// Module is the Go module name in go.mod.
	Module string
	// Files maps a file name to its content.
	Files map[string]string
	// Requires are go.mod requirement lines the project needs beyond Bubble Tea,
	// Bubbles and Lip Gloss; GenerateScreens leaves adding them to the caller.
	Requires []string
	// Notes list what the generator could not do, one line each.
	Notes []string
}

// generatedSuffix ends the name of a generated page, component or master file
// in the slice layout (<page>.cuppa.go, as .pb.go is protobuf's): hand-written
// files sit beside them, and the name says which are Cuppa's. It is valid Go
// and is not a GOOS, GOARCH or _test suffix.
const generatedSuffix = ".cuppa.go"
