//go:build windows

package install

import (
	"strings"
	"testing"
)

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
