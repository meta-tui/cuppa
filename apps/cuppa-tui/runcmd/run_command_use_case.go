// Package runcmd is the "cuppa run" command line: say how the application a
// design draws is built and run, so the designer can host it in a pane. The
// block is kept in the design file; a local file beside it overrides it on one
// computer.
package runcmd

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/runhost/local"
)

const usage = `usage:
  cuppa run show <design.cuppa>             print how the design is run, with the local file applied
  cuppa run set  <design.cuppa> [flags]     set it in the design, or in the local file with --local

flags of set (only the ones given change anything):
  --build "go build -o app ./cmd/app"   the command that builds the program ("" removes it)
  --run   "app --demo"                  the program to host
  --env   NAME=value                    an environment variable of both; repeat for more
  --dir   D                             where the program runs (relative to the design's folder)
  --build-dir D                         where the build runs (empty: the same as --dir)
  --size  120x36                        fix the terminal size (0x0: follow the pane)
  --on-save                             build and relaunch when the design is saved
  --local                               write <design>.run.json instead of the design
  --clear                               remove the settings

Commands are a program and its arguments, not shell lines: no &&, pipes or variables.
`

// Run executes "cuppa run" with the arguments after the word run and returns
// the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "show":
		err = show(args[1], stdout)
	case "set":
		err = set(args[1], args[2:], stdout)
	default:
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa run:", err)
		return 1
	}
	return 0
}

func show(path string, out io.Writer) error {
	doc, err := disk.Load(path)
	if err != nil {
		return err
	}
	s, err := local.Load(path, doc.Run)
	if err != nil {
		return err
	}
	if !s.IsSet() {
		_, _ = fmt.Fprintln(out, "nothing to run is set")
		return nil
	}
	_, _ = fmt.Fprintf(out, "build     %s\n", strings.Join(s.Build, " "))
	_, _ = fmt.Fprintf(out, "run       %s\n", strings.Join(s.Run, " "))
	_, _ = fmt.Fprintf(out, "dir       %s\n", s.Dir)
	_, _ = fmt.Fprintf(out, "build dir %s\n", s.BuildDir)
	_, _ = fmt.Fprintf(out, "size      %dx%d\n", s.Width, s.Height)
	_, _ = fmt.Fprintf(out, "on save   %t\n", s.OnSave)
	names := make([]string, 0, len(s.Env))
	for n := range s.Env {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		_, _ = fmt.Fprintf(out, "env       %s=%s\n", n, s.Env[n])
	}
	return nil
}

// envFlags collects --env NAME=value.
type envFlags map[string]string

func (e envFlags) String() string { return "" }

func (e envFlags) Set(v string) error {
	name, value, ok := strings.Cut(v, "=")
	if !ok || strings.TrimSpace(name) == "" {
		return fmt.Errorf("%q is not NAME=value", v)
	}
	e[strings.TrimSpace(name)] = value
	return nil
}

func set(path string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("run set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	build := fs.String("build", "", "")
	run := fs.String("run", "", "")
	dir := fs.String("dir", "", "")
	buildDir := fs.String("build-dir", "", "")
	size := fs.String("size", "", "")
	onSave := fs.Bool("on-save", false, "")
	localFile := fs.Bool("local", false, "")
	wipe := fs.Bool("clear", false, "")
	env := envFlags{}
	fs.Var(env, "env", "")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w\n%s", err, usage)
	}
	given := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { given[f.Name] = true })

	var buildArgv, runArgv []string
	var err error
	if given["build"] {
		if buildArgv, err = splitCommand(*build); err != nil {
			return err
		}
	}
	if given["run"] {
		if runArgv, err = splitCommand(*run); err != nil {
			return err
		}
		if len(runArgv) == 0 {
			return fmt.Errorf("--run needs a program (use --clear to remove the settings)")
		}
	}
	width, height, err := parseSize(*size, given["size"])
	if err != nil {
		return err
	}

	if *localFile {
		return setLocal(path, out, localChange{
			clear: *wipe, given: given, build: buildArgv, run: runArgv, env: env,
			dir: *dir, buildDir: *buildDir, width: width, height: height, onSave: *onSave,
		})
	}
	doc, err := disk.Load(path)
	if err != nil {
		return err
	}
	s := doc.Run
	if *wipe {
		s = design.RunSettings{}
	}
	if given["build"] {
		s.Build = buildArgv
	}
	if given["run"] {
		s.Run = runArgv
	}
	if len(env) > 0 {
		merged := map[string]string{}
		for k, v := range s.Env {
			merged[k] = v
		}
		for k, v := range env {
			merged[k] = v
		}
		s.Env = merged
	}
	if given["dir"] {
		s.Dir = *dir
	}
	if given["build-dir"] {
		s.BuildDir = *buildDir
	}
	if given["size"] {
		s.Width, s.Height = width, height
	}
	if given["on-save"] {
		s.OnSave = *onSave
	}
	s = s.Clean()
	if !*wipe && !s.IsSet() {
		return fmt.Errorf("the design has no program to run: give --run")
	}
	doc.Run = s
	if _, err := disk.Save(path, doc); err != nil {
		return err
	}
	if s.IsSet() {
		_, _ = fmt.Fprintf(out, "%s runs %s\n", path, strings.Join(s.Run, " "))
	} else {
		_, _ = fmt.Fprintf(out, "%s has no run settings\n", path)
	}
	return nil
}

type localChange struct {
	clear         bool
	given         map[string]bool
	build, run    []string
	env           map[string]string
	dir, buildDir string
	width, height int
	onSave        bool
}

func setLocal(path string, out io.Writer, c localChange) error {
	o, err := local.Read(path)
	if err != nil {
		return err
	}
	if c.clear {
		o = local.Override{}
	}
	if c.given["build"] {
		o.Build = c.build
	}
	if c.given["run"] {
		o.Run = c.run
	}
	if len(c.env) > 0 {
		merged := map[string]string{}
		for k, v := range o.Env {
			merged[k] = v
		}
		for k, v := range c.env {
			merged[k] = v
		}
		o.Env = merged
	}
	if c.given["dir"] {
		o.Dir = &c.dir
	}
	if c.given["build-dir"] {
		o.BuildDir = &c.buildDir
	}
	if c.given["size"] {
		o.Width, o.Height = &c.width, &c.height
	}
	if c.given["on-save"] {
		o.OnSave = &c.onSave
	}
	if err := local.Save(path, o); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "wrote %s (keep it out of version control)\n", local.PathOf(path))
	return nil
}

func parseSize(v string, given bool) (w, h int, err error) {
	if !given {
		return 0, 0, nil
	}
	ws, hs, ok := strings.Cut(strings.ToLower(v), "x")
	if !ok {
		return 0, 0, fmt.Errorf("--size is WxH, like 120x36")
	}
	if w, err = strconv.Atoi(strings.TrimSpace(ws)); err != nil {
		return 0, 0, fmt.Errorf("--size is WxH, like 120x36")
	}
	if h, err = strconv.Atoi(strings.TrimSpace(hs)); err != nil {
		return 0, 0, fmt.Errorf("--size is WxH, like 120x36")
	}
	if w < 0 || h < 0 {
		return 0, 0, fmt.Errorf("--size cannot be negative")
	}
	return w, h, nil
}
