//go:build js

// Package session runs a program in a pseudo-terminal and keeps its screen.
// A browser cannot start programs, so here a session never starts.
package session

import (
	"errors"

	uv "github.com/charmbracelet/ultraviolet"

	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// Session is a program running in a terminal. In the browser there is none.
type Session struct {
	changed chan struct{}
	done    chan struct{}
}

// Start always fails: a browser cannot run the application.
func Start(command.Command, int, int) (*Session, error) {
	return nil, errors.New("session: running the application needs the terminal version of Cuppa")
}

// Changed never receives.
func (s *Session) Changed() <-chan struct{} { return s.changed }

// Done is closed: there is no program.
func (s *Session) Done() <-chan struct{} { return s.done }

// Err is nil.
func (s *Session) Err() error { return nil }

// Render is empty.
func (s *Session) Render() string { return "" }

// Text is empty.
func (s *Session) Text() string { return "" }

// Size is zero.
func (s *Session) Size() (w, h int) { return 0, 0 }

// Resize does nothing.
func (s *Session) Resize(int, int) {}

// SendKey does nothing.
func (s *Session) SendKey(uv.KeyPressEvent) {}

// SendMouse does nothing.
func (s *Session) SendMouse(uv.MouseEvent) {}

// SendText does nothing.
func (s *Session) SendText(string) {}

// Stop does nothing.
func (s *Session) Stop() {}
