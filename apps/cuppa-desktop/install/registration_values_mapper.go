package install

import (
	"encoding/base64"
	"unicode/utf16"
)

// Where Windows looks, relative to the root key (HKCU or HKLM).
const (
	uninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\Cuppa`
	classesKey   = `Software\Classes`
	// progID names what a .cuppa file is, and extension the file type.
	progID    = "Cuppa.Design"
	extension = ".cuppa"
)

// registryValue is one string value: the key under the root, the name of the
// value ("" is the key's default) and the text.
type registryValue struct {
	Key, Name, Value string
}

// uninstallValues describes the program installed at t for Settings > Apps.
// The uninstall command is the program itself with --uninstall, so Windows
// needs nothing else from the install.
func uninstallValues(t target, version string) []registryValue {
	return []registryValue{
		{uninstallKey, "DisplayName", "Cuppa"},
		{uninstallKey, "DisplayVersion", version},
		{uninstallKey, "Publisher", "Cuppa"},
		{uninstallKey, "InstallLocation", t.Folder},
		{uninstallKey, "DisplayIcon", t.Program},
		{uninstallKey, "UninstallString", `"` + t.Program + `" --uninstall`},
	}
}

// uninstallSwitches are the numbers of the entry: Cuppa is neither modified nor
// repaired from Settings, only removed.
var uninstallSwitches = map[string]uint32{"NoModify": 1, "NoRepair": 1}

// associationValues register .cuppa files as Cuppa designs opened by the
// program at path, under Software\Classes of the root.
func associationValues(program string) []registryValue {
	return []registryValue{
		{classesKey + `\` + extension, "", progID},
		{classesKey + `\` + progID, "", "Cuppa design"},
		{classesKey + `\` + progID + `\DefaultIcon`, "", `"` + program + `",0`},
		{classesKey + `\` + progID + `\shell\open\command`, "", `"` + program + `" "%1"`},
	}
}

// associationKeys are the keys an association creates, deepest first, so they
// can be deleted one by one.
func associationKeys() []string {
	return []string{
		classesKey + `\` + progID + `\shell\open\command`,
		classesKey + `\` + progID + `\shell\open`,
		classesKey + `\` + progID + `\shell`,
		classesKey + `\` + progID + `\DefaultIcon`,
		classesKey + `\` + progID,
	}
}

// shortcutScript creates a shortcut. The paths come from the environment, so a
// folder name can never be read as code.
const shortcutScript = `$ErrorActionPreference = 'Stop'
$shortcut = (New-Object -ComObject WScript.Shell).CreateShortcut($env:CUPPA_SHORTCUT)
$shortcut.TargetPath = $env:CUPPA_TARGET
$shortcut.WorkingDirectory = $env:CUPPA_FOLDER
$shortcut.Description = 'Cuppa, a designer for terminal interfaces'
$shortcut.Save()
`

// encodePowerShell is a script as the -EncodedCommand argument reads it: UTF-16
// little endian, base 64.
func encodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, u := range units {
		raw = append(raw, byte(u), byte(u>>8))
	}
	return base64.StdEncoding.EncodeToString(raw)
}
