package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fake is a machine for the offer: it records what was asked, installed and
// started, and answers the questions from a script.
type fake struct {
	base      string
	env       environment
	asked     []string
	told      []string
	installed []target
	started   [][]string
	answers   []answer
	installOK func(t target) error
}

func newFake(t *testing.T) *fake {
	t.Helper()
	f := &fake{base: t.TempDir()}
	f.env = environment{
		GOOS: "windows", Version: "1.2.0",
		Places: places{LocalAppData: filepath.Join(f.base, "Local")},
		AppDir: filepath.Join(f.base, "Config", "Cuppa"),
		Exe:    filepath.Join(f.base, "Downloads", "cuppa-desktop.exe"),
		Args:   []string{"a.cuppa"},
		Ask: func(title, text string, choices []string) answer {
			f.asked = append(f.asked, text)
			if len(f.answers) == 0 {
				return answerLeave
			}
			a := f.answers[0]
			f.answers = f.answers[1:]
			return a
		},
		Tell: func(text string) { f.told = append(f.told, text) },
		Install: func(tg target, exe string) error {
			if f.installOK != nil {
				if err := f.installOK(tg); err != nil {
					return err
				}
			}
			f.installed = append(f.installed, tg)
			return nil
		},
		Start: func(tg target, args []string) error {
			f.started = append(f.started, args)
			return nil
		},
	}
	return f
}

func (f *fake) marker() string {
	data, err := os.ReadFile(filepath.Join(f.env.AppDir, markerName))
	if err != nil {
		return "<none>"
	}
	return string(data)
}

func TestTheFirstStartFromElsewhereOffersToMoveAndRemembersTheVersion(t *testing.T) {
	f := newFake(t)
	f.answers = []answer{answerFirst}
	if got := offer(f.env); got != outcomeMoved {
		t.Fatalf("outcome %v", got)
	}
	if len(f.installed) != 1 || f.installed[0].Program != filepath.Join(f.base, "Local", "Programs", "Cuppa", "Cuppa.exe") {
		t.Fatalf("installed %+v", f.installed)
	}
	if len(f.started) != 1 || strings.Join(f.started[0], "|") != "a.cuppa|--moved-from="+f.env.Exe {
		t.Fatalf("started %v", f.started)
	}
	if !strings.Contains(f.marker(), "installed 1.2.0") {
		t.Fatalf("marker %q", f.marker())
	}
	if !strings.Contains(f.asked[0], filepath.Dir(f.env.Exe)) {
		t.Errorf("the question does not say where it runs from: %q", f.asked[0])
	}
}

func TestADevelopersBuildAndTheCopyAMoveStartedAreNeverAsked(t *testing.T) {
	f := newFake(t)
	f.env.Version = "dev"
	if got := offer(f.env); got != outcomeSkipped || len(f.asked) != 0 {
		t.Fatalf("dev: %v, asked %d", got, len(f.asked))
	}
	f = newFake(t)
	f.env.MovedFrom = f.env.Exe
	if got := offer(f.env); got != outcomeSkipped || len(f.asked) != 0 {
		t.Fatalf("moved: %v", got)
	}
}

func TestNotNowThenNeverAskAgainIsRemembered(t *testing.T) {
	f := newFake(t)
	f.answers = []answer{answerLeave, answerFirst} // Not now, Never ask again
	if got := offer(f.env); got != outcomeDeclined {
		t.Fatalf("outcome %v", got)
	}
	if f.marker() != "asked\n" {
		t.Fatalf("marker %q", f.marker())
	}
	// The next start asks nothing: the marker is there, and nothing is installed.
	f.asked = nil
	if got := offer(f.env); got != outcomeSkipped || len(f.asked) != 0 {
		t.Fatalf("second start: %v, asked %v", got, f.asked)
	}
}

func TestNotNowThenAskMeAgainForgetsIt(t *testing.T) {
	f := newFake(t)
	f.answers = []answer{answerLeave, answerLeave} // Not now, Ask me again
	offer(f.env)
	if f.marker() != "<none>" {
		t.Fatalf("marker %q", f.marker())
	}
}

func TestNothingToShowTheQuestionWithLeavesNoRecord(t *testing.T) {
	f := newFake(t)
	f.answers = []answer{answerUnavailable}
	if got := offer(f.env); got != outcomeCannot || f.marker() != "<none>" {
		t.Fatalf("outcome %v, marker %q", got, f.marker())
	}
}

func TestAnotherVersionOffersToReplaceTheInstalledCopyOncePerVersion(t *testing.T) {
	f := newFake(t)
	installedDir := filepath.Join(t.TempDir(), "Programs", "Cuppa")
	if err := os.MkdirAll(installedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	program := filepath.Join(installedDir, "Cuppa.exe")
	if err := os.WriteFile(program, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	f.env.GOOS = "linux" // paths of this machine; the logic does not depend on the system
	ts := []target{{Folder: installedDir, Program: program}}
	if err := os.MkdirAll(f.env.AppDir, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(f.env.AppDir, markerName)
	if err := os.WriteFile(marker, []byte(markerState{InstalledVersion: "1.1.0"}.String()), 0o600); err != nil {
		t.Fatal(err)
	}

	f.answers = []answer{answerLeave}
	if got := updateOffer(f.env, ts, false, parseMarker("asked\ninstalled 1.1.0\n"), marker); got != outcomeDeclined {
		t.Fatalf("not now: %v", got)
	}
	if !strings.Contains(f.marker(), "declined 1.2.0") || !strings.Contains(f.marker(), "installed 1.1.0") {
		t.Fatalf("marker %q", f.marker())
	}
	// The same version is not asked about again.
	f.asked = nil
	if got := updateOffer(f.env, ts, false, parseMarker(f.marker()), marker); got != outcomeSkipped || len(f.asked) != 0 {
		t.Fatalf("same version: %v, asked %d", got, len(f.asked))
	}
	// A newer one is.
	f.env.Version = "1.3.0"
	f.answers = []answer{answerFirst}
	if got := updateOffer(f.env, ts, false, parseMarker(f.marker()), marker); got != outcomeMoved {
		t.Fatalf("newer: %v", got)
	}
	if len(f.installed) != 1 || f.installed[0].Program != program || !strings.Contains(f.marker(), "installed 1.3.0") {
		t.Fatalf("installed %+v, marker %q", f.installed, f.marker())
	}
}

func TestAnOlderVersionDoesNotReplaceANewerInstalledOne(t *testing.T) {
	if shouldOfferUpdate(updateSituation{Version: "1.0.0", HasInstalledCopy: true, Marker: markerState{InstalledVersion: "1.2.0"}}) {
		t.Fatal("an older version would replace a newer one")
	}
	if !shouldOfferUpdate(updateSituation{Version: "1.3.0", HasInstalledCopy: true, Marker: markerState{InstalledVersion: "1.2.0"}}) {
		t.Fatal("a newer version is not offered")
	}
	if shouldOfferUpdate(updateSituation{Version: "1.3.0", HasInstalledCopy: false, Marker: markerState{}}) {
		t.Fatal("a person who never moved it is asked to update")
	}
}

func TestWhenInstallingForEveryoneFailsTheOwnFolderIsOffered(t *testing.T) {
	f := newFake(t)
	f.env.Places.ProgramFiles = filepath.Join(f.base, "ProgramFiles")
	f.installOK = func(tg target) error {
		if tg.Everyone {
			return errors.New("access denied")
		}
		return nil
	}
	f.answers = []answer{answerFirst, answerFirst} // For everyone, then Install just for me
	if got := offer(f.env); got != outcomeMoved {
		t.Fatalf("outcome %v, told %v", got, f.told)
	}
	if len(f.installed) != 1 || f.installed[0].Everyone {
		t.Fatalf("installed %+v", f.installed)
	}
}

func TestAProgramAlreadyInItsFolderIsNotAskedToMove(t *testing.T) {
	f := newFake(t)
	f.env.Exe = filepath.Join(f.base, "Local", "Programs", "Cuppa", "Cuppa.exe")
	if got := offer(f.env); got != outcomeSkipped || len(f.asked) != 0 {
		t.Fatalf("outcome %v", got)
	}
	f.env.Force = true
	if got := offer(f.env); got != outcomeAlreadyThere {
		t.Fatalf("forced: %v", got)
	}
}
