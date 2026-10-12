// Package command turns a design's run block into the commands to start.
package command

// Command is one program to start: what to run, where and with what
// environment.
type Command struct {
	// Argv is the program and its arguments.
	Argv []string
	// Dir is the working directory, absolute.
	Dir string
	// Env is the environment, as NAME=value, on top of the process's own.
	Env []string
}

// Plan is what saving the design starts: an optional build, then the program.
type Plan struct {
	// Build is nil when there is nothing to build.
	Build *Command
	Run   Command
	// Width and Height fix the program's terminal; zero follows the pane.
	Width, Height int
	// OnSave says to build and relaunch when the design is saved.
	OnSave bool
}
