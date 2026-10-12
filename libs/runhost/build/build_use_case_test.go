package build

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// TestMain makes the test binary a build tool: it prints a line and exits with
// the code in RUNHOST_EXIT.
func TestMain(m *testing.M) {
	if code := os.Getenv("RUNHOST_EXIT"); code != "" {
		fmt.Println("compiling...")
		fmt.Fprintln(os.Stderr, "warning: stderr too")
		if code != "0" {
			os.Exit(2)
		}
		return
	}
	os.Exit(m.Run())
}

func exe(t *testing.T) string {
	t.Helper()
	p, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestABuildReturnsWhatItPrinted(t *testing.T) {
	out, err := Run(context.Background(), command.Command{Argv: []string{exe(t)}, Dir: t.TempDir(), Env: []string{"RUNHOST_EXIT=0"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "compiling...") || !strings.Contains(out, "stderr too") {
		t.Fatalf("output %q", out)
	}
}

func TestAFailedBuildReturnsItsOutputAndAnError(t *testing.T) {
	out, err := Run(context.Background(), command.Command{Argv: []string{exe(t)}, Dir: t.TempDir(), Env: []string{"RUNHOST_EXIT=1"}})
	if err == nil {
		t.Fatal("a failed build is a success")
	}
	if !strings.Contains(out, "compiling...") {
		t.Fatalf("output %q", out)
	}
}

func TestACommandThatDoesNotExistIsAnError(t *testing.T) {
	if _, err := Run(context.Background(), command.Command{Argv: []string{"no-such-program-cuppa"}, Dir: t.TempDir()}); err == nil {
		t.Fatal("no error")
	}
}
