// Package install lets the desktop app put itself where programs live, the way
// MVD does: it offers to move out of the folder it was downloaded to, lists
// itself in Settings > Apps, registers .cuppa files, adds a Start menu entry
// and offers to replace an older installed copy with the version that runs.
// Only Windows installs itself; elsewhere the app keeps running from where it
// is.
package install

import (
	"path/filepath"
	"strings"
)

// folderName is the folder the program lives in, and programName the program.
const (
	folderName  = "Cuppa"
	programName = "Cuppa.exe"
)

// kind says how the program is put in place, which differs by system.
type kind int

const (
	// kindSystem is Program Files, offered only when the app already runs as an
	// administrator, which is when a plain copy works.
	kindSystem kind = iota
	// kindUser is the person's own Programs folder.
	kindUser
)

// target is one place the program can be moved to.
type target struct {
	Kind kind
	// Everyone is true for a place that serves every account.
	Everyone bool
	// Folder is where it goes and Program the path of the program there.
	Folder  string
	Program string
}

// places are the folders of the machine the targets are built from. Either may
// be empty, which leaves out the target that needs it. ProgramFiles is only
// filled in when the app may write there.
type places struct {
	ProgramFiles string
	LocalAppData string
}

// targets lists where the program belongs on goos: the place for everyone
// first, the person's own place after it. Only Windows has one.
func targets(goos string, p places) []target {
	if goos != "windows" {
		return nil
	}
	var out []target
	if p.ProgramFiles != "" {
		folder := filepath.Join(p.ProgramFiles, folderName)
		out = append(out, target{Kind: kindSystem, Everyone: true, Folder: folder, Program: filepath.Join(folder, programName)})
	}
	if p.LocalAppData != "" {
		folder := filepath.Join(p.LocalAppData, "Programs", folderName)
		out = append(out, target{Kind: kindUser, Folder: folder, Program: filepath.Join(folder, programName)})
	}
	return out
}

// samePlace reports whether two paths are the same; Windows paths are not case
// sensitive.
func samePlace(goos, a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if goos == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// isInstalled reports whether the program at exe is in one of the targets.
func isInstalled(goos, exe string, ts []target) bool {
	for _, t := range ts {
		if samePlace(goos, filepath.Dir(exe), filepath.Dir(t.Program)) {
			return true
		}
	}
	return false
}
