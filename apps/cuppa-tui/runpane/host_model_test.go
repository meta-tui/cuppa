package runpane

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// TestMain makes the test binary the program being hosted, or the build tool.
func TestMain(m *testing.M) {
	switch os.Getenv("RUNPANE_CHILD") {
	case "app":
		say("hello from the app")
		buf := make([]byte, 64)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			say("typed:" + strings.TrimSpace(string(buf[:n])))
		}
	case "build-fails":
		say("main.go:3: undefined: nothing")
		os.Exit(2)
	}
	os.Exit(m.Run())
}

func say(text string) { _, _ = os.Stdout.WriteString(text + "\r\n") }

func exe(t *testing.T) string {
	t.Helper()
	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// text is the grid as lines.
func text(g *grid.Grid) string {
	var b strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.At(x, y)
			if c.Ch == 0 {
				b.WriteRune(' ')
			} else {
				b.WriteRune(c.Ch)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// drive feeds the messages the commands produce back to the model until the
// grid shows want.
func drive(t *testing.T, m *Model, cmd tea.Cmd, want string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(text(m.Grid()), want) {
			return
		}
		if cmd == nil {
			break
		}
		cmd = m.Update(cmd())
	}
	t.Fatalf("the pane never showed %q (state %d):\n%s", want, m.State(), text(m.Grid()))
}

func appPlan(t *testing.T) command.Plan {
	t.Helper()
	return command.Plan{Run: command.Command{Argv: []string{exe(t)}, Dir: t.TempDir(), Env: []string{"RUNPANE_CHILD=app"}}}
}

func TestThePaneShowsTheProgramAndGivesItTheKeys(t *testing.T) {
	m, cmd := Start(appPlan(t), nil, 50, 8)
	t.Cleanup(m.Stop)
	if m.State() != Running {
		t.Skipf("no pseudo-terminal here: %s", m.output)
	}
	drive(t, m, cmd, "hello from the app")
	m.Key(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m.Key(tea.KeyPressMsg{Code: tea.KeyEnter})
	drive(t, m, m.wait(), "typed:x")
}

func TestAFailedBuildShowsWhatTheBuildSaid(t *testing.T) {
	p := appPlan(t)
	p.Build = &command.Command{Argv: []string{exe(t)}, Dir: p.Run.Dir, Env: []string{"RUNPANE_CHILD=build-fails"}}
	m, cmd := Start(p, nil, 60, 6)
	t.Cleanup(m.Stop)
	drive(t, m, cmd, "undefined: nothing")
	if m.State() != BuildFailed {
		t.Fatalf("state %d", m.State())
	}
	m.Key(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m.Mouse(pointer.Event{X: 1, Y: 1, Phase: pointer.Down, Left: true})
}

func TestARestartRunsTheProgramAgainAndIgnoresLateMessages(t *testing.T) {
	m, cmd := Start(appPlan(t), nil, 50, 8)
	t.Cleanup(m.Stop)
	if m.State() != Running {
		t.Skipf("no pseudo-terminal here: %s", m.output)
	}
	drive(t, m, cmd, "hello from the app")
	stale := ChangedMsg{Gen: m.gen}
	again := m.Restart(appPlan(t), nil, 50, 8)
	if got := m.Update(stale); got != nil {
		t.Fatal("a message from the old run was taken")
	}
	drive(t, m, again, "hello from the app")
}

func TestAPlanThatFixesTheSizeKeepsIt(t *testing.T) {
	p := appPlan(t)
	p.Width, p.Height = 70, 9
	m, _ := Start(p, nil, 50, 8)
	t.Cleanup(m.Stop)
	m.Resize(120, 40)
	if w, h := m.Size(); w != 70 || h != 9 {
		t.Fatalf("size %dx%d", w, h)
	}
}
