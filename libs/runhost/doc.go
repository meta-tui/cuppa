// Package runhost hosts the application a design draws, so the designer can
// show exactly what it prints next to the editor: the program runs in a
// pseudo-terminal and its output is drawn by a terminal emulator. Nothing is
// written in the application. Its code lives in the slice packages below this
// directory:
//
//	command  what to build and run, from the design's run block
//	build    run the build command and keep what it printed
//	session  the program in a pseudo-terminal: its screen, keys and mouse in
//
// It is headless: it imports no Bubble Tea and no Lip Gloss, only the
// terminal emulator and the pty packages.
package runhost
