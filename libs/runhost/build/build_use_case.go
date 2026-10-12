// Package build runs the build command of a design's run block.
package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/runhost/command"
)

// Run runs the command and returns what it printed, standard output and
// error together. The error is set when it could not start or exited with a
// failure; the output is what to show the person then.
func Run(ctx context.Context, c command.Command) (string, error) {
	if len(c.Argv) == 0 {
		return "", fmt.Errorf("build: no command")
	}
	cmd := exec.CommandContext(ctx, c.Argv[0], c.Argv[1:]...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), c.Env...)
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\r\n ")
	if err != nil {
		return text, fmt.Errorf("build: %s: %w", c.Argv[0], err)
	}
	return text, nil
}
