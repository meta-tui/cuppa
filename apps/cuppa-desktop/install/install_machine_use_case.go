package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// Startup is what a start of the program does about installing, on the real
// machine: with force (--install) it asks to move whatever was answered before;
// otherwise it offers once, and offers to update an installed copy of another
// version. It reports true when the program was moved and started from its new
// place, and the caller should exit. movedFrom is the copy a move left behind,
// which is removed in the background once that copy has exited.
func Startup(version, movedFrom string, force bool) (moved bool) {
	if movedFrom != "" {
		go CleanUpMovedProgram(movedFrom)
	}
	env, ok := realEnvironment(version, movedFrom, force)
	if !ok {
		return false
	}
	switch offer(env) {
	case outcomeMoved:
		return true
	case outcomeAlreadyThere:
		tell("Cuppa is already installed in its own folder:\n\n" + filepath.Dir(env.Exe))
	case outcomeCannot:
		if force {
			tell("Cuppa cannot move itself on this system.")
		}
	}
	keepRegistered(env)
	return false
}

// keepRegistered writes the registrations again when the program runs from
// its own folder, so a folder that was moved or a registry that was cleaned
// follows. A developer's build is never registered.
func keepRegistered(env environment) {
	if env.Version == "dev" {
		return
	}
	for _, t := range targets(env.GOOS, env.Places) {
		if samePlace(env.GOOS, filepath.Dir(env.Exe), t.Folder) {
			register(t, env.Version)
			return
		}
	}
}

// register lists the program in Settings > Apps, registers .cuppa files and
// adds a Start menu entry. A missing shortcut or entry is a loss of
// convenience, not a reason to undo an install.
func register(t target, version string) {
	_ = makeShortcut(startMenuLink(t.Everyone), t.Program)
	_ = registerUninstall(t, version)
	_ = registerFiles(t)
}

// installOnMachine puts the program in t and registers it.
func installOnMachine(t target, exe, version string) error {
	if err := place(exe, t.Program); err != nil {
		return err
	}
	register(t, version)
	return nil
}

func realEnvironment(version, movedFrom string, force bool) (environment, bool) {
	exe, err := os.Executable()
	if err != nil {
		return environment{}, false
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return environment{}, false
	}
	// Program Files can only be written by an administrator, so it is only a
	// place to move to when the program already runs as one.
	programFiles := ""
	if isElevated() {
		programFiles = os.Getenv("ProgramFiles")
	}
	return environment{
		GOOS: goos, Version: version, MovedFrom: movedFrom, Force: force,
		Places:    places{ProgramFiles: programFiles, LocalAppData: os.Getenv("LOCALAPPDATA")},
		AdminHint: goos == "windows" && !isElevated(),
		AppDir:    filepath.Join(configDir, folderName),
		Exe:       exe,
		Args:      userArgs(os.Args[1:]),
		Ask:       ask,
		Tell:      tell,
		Install:   func(t target, exe string) error { return installOnMachine(t, exe, version) },
		Start:     startInstalled,
	}, true
}

// userArgs are the arguments the copy is started with: the person's own, not
// the ones that steer the install.
func userArgs(args []string) []string {
	var out []string
	for _, a := range args {
		if a == "--install" || len(a) > len("--moved-from=") && a[:len("--moved-from=")] == "--moved-from=" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// startMenuLink is the Cuppa shortcut of the person's Start menu, or of every
// account's.
func startMenuLink(everyone bool) string {
	if everyone {
		return filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs", "Cuppa.lnk")
	}
	return filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Cuppa.lnk")
}

// Uninstall asks, and then removes what the install wrote: the program and its
// folder, the Start menu shortcuts and the registrations. The person's designs
// and settings are not touched. It reports true when the program was removed,
// and the caller should exit.
func Uninstall() bool {
	env, ok := realEnvironment("", "", false)
	if !ok {
		return false
	}
	ts := targets(env.GOOS, places{ProgramFiles: os.Getenv("ProgramFiles"), LocalAppData: env.Places.LocalAppData})
	if ask("Cuppa", "Remove Cuppa from this computer?\n\nYour designs and settings are not touched.", []string{"Remove it", "Cancel"}) != answerFirst {
		return false
	}
	var left []string
	_ = removeRegistrations()
	for _, everyone := range []bool{false, true} {
		_ = os.Remove(startMenuLink(everyone))
	}
	for _, t := range ts {
		if samePlace(env.GOOS, filepath.Dir(env.Exe), t.Folder) {
			leftover, err := removeRunningProgram(env.Exe)
			switch {
			case err != nil:
				left = append(left, fmt.Sprintf("%s (%v)", env.Exe, err))
			case leftover != "":
				left = append(left, leftover)
			}
		} else if _, err := os.Stat(t.Program); err == nil {
			if err := os.Remove(t.Program); err != nil {
				left = append(left, fmt.Sprintf("%s (%v)", t.Program, err))
			}
		}
		_ = os.Remove(t.Program + ".part")
		_ = os.Remove(t.Folder) // only when nothing else is in it
	}
	if len(left) > 0 {
		tell("Cuppa was removed from Settings and the Start menu, but these are still there:\n\n" + joinLines(left))
	}
	return true
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}
