package install

import (
	"strconv"
	"strings"
)

// markerState is what the "install-offered" file in the app-data folder
// remembers: that the person was asked to move the program, which version was
// put in place, and which version they said "Not now" to updating it with, so
// the same version is not asked about twice. A file that only says "asked"
// leaves both versions empty.
type markerState struct {
	InstalledVersion string
	DeclinedUpdate   string
}

func parseMarker(text string) markerState {
	var s markerState
	for _, line := range strings.Split(text, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found {
			continue
		}
		switch key {
		case "installed":
			s.InstalledVersion = strings.TrimSpace(value)
		case "declined":
			s.DeclinedUpdate = strings.TrimSpace(value)
		}
	}
	return s
}

func (s markerState) String() string {
	text := "asked\n"
	if s.InstalledVersion != "" {
		text += "installed " + s.InstalledVersion + "\n"
	}
	if s.DeclinedUpdate != "" {
		text += "declined " + s.DeclinedUpdate + "\n"
	}
	return text
}

// versionOlder reports whether version a is older than b, for "major.minor.patch".
// A version that cannot be read is never called older, so the person is asked
// rather than left on a copy that is out of date.
func versionOlder(a, b string) bool {
	pa, okA := versionParts(a)
	pb, okB := versionParts(b)
	if !okA || !okB {
		return false
	}
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func versionParts(version string) (parts [3]int, ok bool) {
	pieces := strings.Split(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".")
	if len(pieces) != 3 {
		return parts, false
	}
	for i, piece := range pieces {
		n, err := strconv.Atoi(piece)
		if err != nil {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

// moveSituation is everything that decides whether to offer moving the program.
type moveSituation struct {
	// Version is "dev" for a build made on a developer's machine.
	Version string
	// MovedFrom is set on the copy that a move has just started.
	MovedFrom string
	HasTarget bool
	Installed bool
}

// shouldOfferMove: ask the first time the program is started from somewhere it
// does not belong, and never from a developer's build or the copy a move just
// started.
func shouldOfferMove(s moveSituation) bool {
	switch {
	case !s.HasTarget, s.Version == "dev", s.MovedFrom != "", s.Installed:
		return false
	}
	return true
}

// updateSituation is everything that decides whether to offer replacing the
// installed copy with the program that runs.
type updateSituation struct {
	Version          string
	MovedFrom        string
	Installed        bool
	HasInstalledCopy bool
	Marker           markerState
}

// shouldOfferUpdate: a copy is installed, this is not it, and this is not the
// same version, nor an older one, nor one the person already said "Not now" to.
func shouldOfferUpdate(s updateSituation) bool {
	switch {
	case s.Version == "dev", s.MovedFrom != "", s.Installed, !s.HasInstalledCopy:
		return false
	case s.Marker.DeclinedUpdate == s.Version:
		return false
	case s.Marker.InstalledVersion == s.Version:
		return false
	case s.Marker.InstalledVersion != "" && versionOlder(s.Version, s.Marker.InstalledVersion):
		return false
	}
	return true
}
