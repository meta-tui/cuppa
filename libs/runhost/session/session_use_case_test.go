package session

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// TestMain makes the test binary its own program to host: with the variable
// set it is a small line-echoing program instead of the tests.
func TestMain(m *testing.M) {
	if os.Getenv("RUNHOST_CHILD") == "1" {
		child()
		return
	}
	os.Exit(m.Run())
}

func child() {
	say("ready")
	buf := make([]byte, 256)
	line := ""
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return
		}
		line += string(buf[:n])
		if strings.ContainsAny(line, "\r\n") {
			text := strings.TrimSpace(line)
			if text == "quit" {
				say("bye")
				return
			}
			say("got:" + text)
			line = ""
		}
	}
}

func startChild(t *testing.T) *Session {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Start(command.Command{Argv: []string{exe}, Dir: t.TempDir(), Env: []string{"RUNHOST_CHILD=1"}}, 60, 10)
	if err != nil {
		t.Skipf("no pseudo-terminal here: %v", err)
	}
	t.Cleanup(s.Stop)
	return s
}

func waitFor(t *testing.T, s *Session, want string) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		if strings.Contains(s.Text(), want) {
			return
		}
		select {
		case <-s.Changed():
		case <-time.After(100 * time.Millisecond):
		case <-deadline:
			t.Fatalf("the screen never showed %q:\n%s", want, s.Text())
		}
	}
}

func TestTheScreenShowsWhatTheProgramPrintsAndKeysReachIt(t *testing.T) {
	s := startChild(t)
	waitFor(t, s, "ready")
	s.SendText("abc\r")
	waitFor(t, s, "got:abc")
}

func TestTheSessionEndsWhenTheProgramDoes(t *testing.T) {
	s := startChild(t)
	waitFor(t, s, "ready")
	s.SendText("quit\r")
	select {
	case <-s.Done():
	case <-time.After(10 * time.Second):
		t.Fatalf("the program did not end:\n%s", s.Text())
	}
	waitFor(t, s, "bye")
}

func TestTheTerminalTakesAnotherSize(t *testing.T) {
	s := startChild(t)
	waitFor(t, s, "ready")
	s.Resize(40, 12)
	if w, h := s.Size(); w != 40 || h != 12 {
		t.Fatalf("size %dx%d", w, h)
	}
}

func TestStoppingTwiceIsFine(t *testing.T) {
	s := startChild(t)
	waitFor(t, s, "ready")
	s.Stop()
	s.Stop()
}

func say(text string) { _, _ = os.Stdout.WriteString(text + "\r\n") }
