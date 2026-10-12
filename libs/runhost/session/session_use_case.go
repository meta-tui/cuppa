// Package session runs a program in a pseudo-terminal and keeps its screen.
package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"

	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// Session is a program running in a terminal of w by h cells. What it prints
// is drawn by an emulator, and keys and mouse events go to it as a terminal
// sends them. It is safe for use from several goroutines.
type Session struct {
	pty xpty.Pty
	cmd *exec.Cmd

	mu sync.Mutex
	em *vt.Emulator

	changed chan struct{}
	done    chan struct{}
	exit    error
	stop    context.CancelFunc
}

// Start runs c in a terminal of w by h cells.
func Start(c command.Command, w, h int) (*Session, error) {
	if len(c.Argv) == 0 {
		return nil, fmt.Errorf("session: no command")
	}
	w, h = max(w, 1), max(h, 1)
	p, err := xpty.NewPty(w, h)
	if err != nil {
		return nil, fmt.Errorf("session: pseudo-terminal: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, c.Argv[0], c.Argv[1:]...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")
	cmd.Env = append(cmd.Env, c.Env...)
	if err := p.Start(cmd); err != nil {
		cancel()
		_ = p.Close()
		return nil, fmt.Errorf("session: start %s: %w", c.Argv[0], err)
	}
	s := &Session{
		pty:     p,
		cmd:     cmd,
		em:      vt.NewEmulator(w, h),
		changed: make(chan struct{}, 1),
		done:    make(chan struct{}),
		stop:    cancel,
	}
	go s.pump()
	go s.answer()
	go func() {
		s.exit = xpty.WaitProcess(ctx, cmd)
		close(s.done)
	}()
	return s, nil
}

// pump draws what the program prints.
func (s *Session) pump() {
	buf := make([]byte, 32*1024)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.mu.Lock()
			_, _ = s.em.Write(buf[:n])
			s.mu.Unlock()
			select {
			case s.changed <- struct{}{}:
			default:
			}
		}
		if err != nil {
			return
		}
	}
}

// answer sends to the program what the emulator says to it: keys, mouse
// events, and the replies a terminal gives to a query.
func (s *Session) answer() {
	buf := make([]byte, 4096)
	for {
		n, err := s.em.Read(buf)
		if n > 0 {
			if _, werr := s.pty.Write(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// Changed receives a value after the program printed something. A slow reader
// loses nothing: the next screen shows everything.
func (s *Session) Changed() <-chan struct{} { return s.changed }

// Done is closed when the program has ended.
func (s *Session) Done() <-chan struct{} { return s.done }

// Err is how the program ended, once Done is closed.
func (s *Session) Err() error {
	select {
	case <-s.done:
		return s.exit
	default:
		return nil
	}
}

// Render is the screen as text with its colours, one line per row.
func (s *Session) Render() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.em.Render()
}

// Text is the screen as plain text, one line per row.
func (s *Session) Text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.TrimRight(s.em.String(), "\n")
}

// Size is the size of the terminal in cells.
func (s *Session) Size() (w, h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.em.Width(), s.em.Height()
}

// Resize sets the size of the terminal, and tells the program.
func (s *Session) Resize(w, h int) {
	w, h = max(w, 1), max(h, 1)
	s.mu.Lock()
	same := s.em.Width() == w && s.em.Height() == h
	if !same {
		s.em.Resize(w, h)
	}
	s.mu.Unlock()
	if !same {
		_ = s.pty.Resize(w, h)
	}
}

// SendKey gives the program a key, as the terminal would have sent it.
func (s *Session) SendKey(k uv.KeyPressEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.em.SendKey(k)
}

// SendMouse gives the program a mouse event, in cells of its terminal. The
// program sees it only when it has asked for the mouse.
func (s *Session) SendMouse(m uv.MouseEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.em.SendMouse(m)
}

// SendText gives the program text, as a paste.
func (s *Session) SendText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.em.SendText(text)
}

// Stop ends the program and releases the terminal. It is safe to call twice.
func (s *Session) Stop() {
	s.stop()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.pty.Close()
	<-s.done
	s.mu.Lock()
	_ = s.em.Close()
	s.mu.Unlock()
}
