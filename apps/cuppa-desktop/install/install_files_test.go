package install

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlaceReplacesAnOldProgramAndLeavesNoPartialFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new.exe")
	dst := filepath.Join(dir, "Programs", "Cuppa", "Cuppa.exe")
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := place(src, dst); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(dst); string(data) != "new" {
		t.Fatalf("dst %q", data)
	}
	if _, err := os.Stat(dst + ".part"); err == nil {
		t.Error("a partial file is left")
	}
}

func TestTheMovedProgramIsRemovedOnceItHasExited(t *testing.T) {
	calls := 0
	remove := func(string) error {
		calls++
		if calls < 3 {
			return errors.New("in use")
		}
		return nil
	}
	if !removeAfterExit("x.exe", 10, time.Millisecond, remove, func(time.Duration) {}) || calls != 3 {
		t.Fatalf("calls %d", calls)
	}
	if !removeAfterExit("x.exe", 3, time.Millisecond, func(string) error { return fs.ErrNotExist }, func(time.Duration) {}) {
		t.Fatal("a program that is gone is not done")
	}
	if removeAfterExit("x.exe", 2, time.Millisecond, func(string) error { return errors.New("in use") }, func(time.Duration) {}) {
		t.Fatal("a program that cannot be removed was reported removed")
	}
}

func TestTheRegistrationsAreWhatWindowsReads(t *testing.T) {
	tg := target{Folder: `C:\P\Cuppa`, Program: `C:\P\Cuppa\Cuppa.exe`}
	got := map[string]string{}
	for _, v := range uninstallValues(tg, "1.2.0") {
		got[v.Name] = v.Value
	}
	if got["UninstallString"] != `"C:\P\Cuppa\Cuppa.exe" --uninstall` || got["DisplayVersion"] != "1.2.0" || got["InstallLocation"] != `C:\P\Cuppa` {
		t.Errorf("uninstall values %v", got)
	}
	var open string
	var ext string
	for _, v := range associationValues(tg.Program) {
		if strings.HasSuffix(v.Key, `shell\open\command`) {
			open = v.Value
		}
		if v.Key == `Software\Classes\.cuppa` {
			ext = v.Value
		}
	}
	if open != `"C:\P\Cuppa\Cuppa.exe" "%1"` || ext != progID {
		t.Errorf("association: open %q, ext %q", open, ext)
	}
	keys := associationKeys()
	if keys[len(keys)-1] != `Software\Classes\Cuppa.Design` || !strings.HasSuffix(keys[0], `command`) {
		t.Errorf("keys %v", keys)
	}
}

func TestTheShortcutScriptIsEncodedForPowerShell(t *testing.T) {
	enc := encodePowerShell("ab")
	if enc != "YQBiAA==" { // "a\x00b\x00"
		t.Fatalf("encoded %q", enc)
	}
}

func TestTheArgumentsOfTheCopyAreThePersonsOwn(t *testing.T) {
	got := userArgs([]string{"--install", "--moved-from=C:\\x.exe", `C:\a.cuppa`})
	if len(got) != 1 || got[0] != `C:\a.cuppa` {
		t.Fatalf("args %v", got)
	}
}

func TestVersionsAreComparedByNumber(t *testing.T) {
	if !versionOlder("0.0.9", "0.0.10") || versionOlder("0.0.10", "0.0.9") || versionOlder("1.0.0", "1.0.0") {
		t.Fatal("numbers are compared as text")
	}
	if versionOlder("dev", "1.0.0") || versionOlder("1.0.0", "x") {
		t.Fatal("an unreadable version is called older")
	}
	s := parseMarker(markerState{InstalledVersion: "1.0.0", DeclinedUpdate: "1.1.0"}.String())
	if s.InstalledVersion != "1.0.0" || s.DeclinedUpdate != "1.1.0" {
		t.Fatalf("marker %+v", s)
	}
}
