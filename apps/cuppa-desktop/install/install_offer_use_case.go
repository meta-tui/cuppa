package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// answer is the choice a question got.
type answer int

const (
	// answerFirst, answerSecond: the first and second choices offered.
	answerFirst answer = iota
	answerSecond
	// answerLeave: the last choice, or the box was closed.
	answerLeave
	// answerUnavailable: nothing could be shown, so nothing was asked.
	answerUnavailable
)

// markerName is the file in the app-data folder that records the person was
// asked to move the program, whatever they answered.
const markerName = "install-offered"

// environment is what moving the program needs from the machine, so that every
// branch can be tested without a screen or a real installation.
type environment struct {
	GOOS    string
	Version string
	// MovedFrom is set on the copy a move has just started.
	MovedFrom string
	Places    places
	// AdminHint adds to the question how to install for everyone.
	AdminHint bool
	// AppDir is where the marker is kept.
	AppDir string
	// Exe is the running program and Args the arguments it was started with.
	Exe  string
	Args []string
	// Force is true when the person asked to move it (--install): they are asked
	// whatever they answered before and nothing is remembered.
	Force bool

	Ask     func(title, text string, choices []string) answer
	Tell    func(text string)
	Install func(t target, exe string) error
	Start   func(t target, args []string) error
}

// outcome is how an offer went.
type outcome int

const (
	// outcomeSkipped: the offer does not apply, so nobody was asked.
	outcomeSkipped outcome = iota
	// outcomeMoved: installed and the copy started; the caller should exit.
	outcomeMoved
	outcomeDeclined
	outcomeAlreadyThere
	outcomeCannot
)

// offer asks the person, the first time the program is started from somewhere
// it does not belong, whether to move it, and where. If they agree it installs
// itself and starts the copy, which removes this one once it has exited. After
// that, a different version started from elsewhere is offered to replace the
// installed one.
//
// The question is recorded as asked before it is put, so however the program
// ends it is not asked in a loop; the record is taken back when the person says
// to ask again or when nothing could show the question.
func offer(env environment) outcome {
	ts := targets(env.GOOS, env.Places)
	if len(ts) == 0 {
		return outcomeCannot
	}
	installed := isInstalled(env.GOOS, env.Exe, ts)
	marker := filepath.Join(env.AppDir, markerName)

	if env.Force {
		if installed {
			return outcomeAlreadyThere
		}
	} else {
		if data, err := os.ReadFile(marker); err == nil {
			return updateOffer(env, ts, installed, parseMarker(string(data)), marker)
		}
		if !shouldOfferMove(moveSituation{Version: env.Version, MovedFrom: env.MovedFrom, HasTarget: true, Installed: installed}) {
			return outcomeSkipped
		}
		if err := os.MkdirAll(env.AppDir, 0o755); err != nil {
			return outcomeSkipped
		}
		if err := os.WriteFile(marker, []byte("asked\n"), 0o600); err != nil {
			return outcomeSkipped
		}
	}

	text, choices := moveQuestion(filepath.Dir(env.Exe), ts, env.AdminHint, env.Force)
	picked := env.Ask("Cuppa", text, choices)
	switch picked {
	case answerUnavailable:
		if !env.Force {
			_ = os.Remove(marker)
		}
		return outcomeCannot
	case answerLeave:
		if !env.Force {
			askAgainLater(env, marker)
		}
		return outcomeDeclined
	}

	chosen := ts[0]
	if picked == answerSecond && len(ts) > 1 {
		chosen = ts[1]
	}
	err := env.Install(chosen, env.Exe)
	if err != nil && chosen.Everyone && len(ts) > 1 {
		fallback := ts[1]
		again := env.Ask("Cuppa", fmt.Sprintf("Cuppa could not be installed for everyone: %v\n\nInstall it just for you instead?\n\n%s", err, fallback.Folder),
			[]string{"Install just for me", "Leave it here"})
		if again != answerFirst {
			return outcomeDeclined
		}
		chosen = fallback
		err = env.Install(chosen, env.Exe)
	}
	if err != nil {
		env.Tell(fmt.Sprintf("Cuppa could not move itself to %s: %v\n\nIt will keep running from where it is.", chosen.Folder, err))
		return outcomeDeclined
	}
	remember(env, marker, markerState{InstalledVersion: env.Version})
	return start(env, chosen)
}

// start runs the installed copy with the arguments this one had, and the place
// it came from, which the copy removes.
func start(env environment, t target) outcome {
	args := append(append([]string{}, env.Args...), "--moved-from="+env.Exe)
	if err := env.Start(t, args); err != nil {
		env.Tell(fmt.Sprintf("Cuppa was installed in %s but could not be started from there: %v\n\nIt will keep running from where it is.", t.Folder, err))
		return outcomeDeclined
	}
	return outcomeMoved
}

func remember(env environment, marker string, s markerState) {
	if env.Force && s.InstalledVersion == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(marker), 0o755)
	_ = os.WriteFile(marker, []byte(s.String()), 0o600)
}

// askAgainLater is the question after "Not now". Only "Never ask again" keeps
// the record; closing the box leaves things as they were and asks again.
func askAgainLater(env environment, marker string) {
	picked := env.Ask("Cuppa",
		"Ask again the next time Cuppa starts?\n\nYou can always move it later: start Cuppa with --install.",
		[]string{"Never ask again", "Ask me again"})
	if picked != answerFirst {
		_ = os.Remove(marker)
	}
}

// updateOffer is the question put when the program is started from somewhere
// other than its own folder after it was moved there once: when a different
// version is installed, replace it with the one that runs.
func updateOffer(env environment, ts []target, installed bool, state markerState, marker string) outcome {
	var present *target
	for i := range ts {
		if _, err := os.Stat(ts[i].Program); err == nil {
			present = &ts[i]
			break
		}
	}
	if !shouldOfferUpdate(updateSituation{Version: env.Version, MovedFrom: env.MovedFrom, Installed: installed, HasInstalledCopy: present != nil, Marker: state}) {
		return outcomeSkipped
	}
	have := "another version"
	if state.InstalledVersion != "" {
		have = "version " + state.InstalledVersion
	}
	text := fmt.Sprintf("Cuppa %s is running from:\n\n%s\n\nYou have %s of Cuppa installed in:\n\n%s\n\nReplace the installed copy with this one?",
		env.Version, filepath.Dir(env.Exe), have, present.Folder)
	switch env.Ask("Cuppa", text, []string{"Update it", "Not now"}) {
	case answerUnavailable:
		return outcomeCannot
	case answerFirst:
	default:
		state.DeclinedUpdate = env.Version
		_ = os.WriteFile(marker, []byte(state.String()), 0o600)
		return outcomeDeclined
	}
	if err := env.Install(*present, env.Exe); err != nil {
		env.Tell(fmt.Sprintf("Cuppa could not update the copy in %s: %v\n\nIf Cuppa is open, close it and start this one again. It will keep running from where it is.", present.Folder, err))
		return outcomeDeclined
	}
	remember(env, marker, markerState{InstalledVersion: env.Version})
	return start(env, *present)
}

// moveQuestion words the question and the choices.
func moveQuestion(from string, ts []target, adminHint, requested bool) (string, []string) {
	text := fmt.Sprintf("Cuppa is running from:\n\n%s\n\n", from)
	footer := "You can also do this later: start Cuppa with --install."
	if requested {
		footer = "You asked for this with --install."
	}
	if len(ts) < 2 {
		text += fmt.Sprintf("Move it to its own folder, where this system keeps programs?\n\n%s\n\n", ts[0].Folder)
		if adminHint {
			text += "To install it for everyone on this computer instead, say No, then start Cuppa as an administrator (right-click it and choose Run as administrator). You will not be asked again, so this is the only time to choose.\n\n"
		}
		return text + footer, []string{"Move it", "Not now"}
	}
	text += "Where should it live?\n\n"
	text += fmt.Sprintf("For everyone on this computer:\n%s\n\n", ts[0].Folder)
	text += fmt.Sprintf("Just for you:\n%s\n\n", ts[1].Folder)
	return text + footer, []string{"For everyone", "Just for me", "Not now"}
}
