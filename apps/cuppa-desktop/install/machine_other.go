//go:build !windows

package install

import "errors"

// Only Windows installs itself: elsewhere the program keeps running from where
// it is, and none of these is ever reached because there is no target.
const goos = "other"

var errUnsupported = errors.New("only Windows installs itself")

func ask(string, string, []string) answer { return answerUnavailable }

func tell(string) {}

func isElevated() bool { return false }

func registerUninstall(target, string) error { return errUnsupported }

func registerFiles(target) error { return errUnsupported }

func removeRegistrations() error { return nil }

func makeShortcut(string, string) error { return errUnsupported }

func startInstalled(target, []string) error { return errUnsupported }

func removeRunningProgram(string) (string, error) { return "", errUnsupported }
