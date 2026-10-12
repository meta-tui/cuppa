// Package runpane hosts the application a design draws in the stage: it
// builds the program, runs it in a pseudo-terminal and draws its screen, and
// passes the keys and mouse the person gives the pane to it.
package runpane

// BuiltMsg says the build step ended; Out is what it printed.
type BuiltMsg struct {
	Gen int
	Out string
	Err error
}

// ChangedMsg says the program printed something or ended.
type ChangedMsg struct{ Gen int }

// State is what the pane is doing.
type State int

// States of the pane.
const (
	// Building is the build step running.
	Building State = iota
	// BuildFailed is a build that ended in an error; its output is shown.
	BuildFailed
	// Running is the program running.
	Running
	// Ended is the program having exited.
	Ended
)
