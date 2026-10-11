//go:build windows

package install

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const goos = "windows"

// What a message box returns for the button that was pressed.
const (
	idYes = 6
	idNo  = 7
)

// ask puts a question in a message box. With two choices it is Yes and No; with
// three, Yes, No and Cancel, and the box says which button means what. Closing
// the box counts as the last choice, leaving things as they are.
func ask(title, text string, choices []string) answer {
	flags := uint32(windows.MB_YESNO)
	if len(choices) == 3 {
		flags = windows.MB_YESNOCANCEL
		text += "\n\nYes: " + choices[0] + "      No: " + choices[1] + "      Cancel: " + choices[2]
	}
	body, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return answerUnavailable
	}
	caption, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return answerUnavailable
	}
	pressed, _ := windows.MessageBox(0, body, caption, flags|windows.MB_ICONQUESTION|windows.MB_SETFOREGROUND)
	switch pressed {
	case idYes:
		return answerFirst
	case idNo:
		if len(choices) == 3 {
			return answerSecond
		}
	}
	return answerLeave
}

// tell shows a problem.
func tell(text string) {
	body, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	caption, _ := windows.UTF16PtrFromString("Cuppa")
	_, _ = windows.MessageBox(0, body, caption, windows.MB_OK|windows.MB_ICONWARNING|windows.MB_SETFOREGROUND)
}

// isElevated reports whether the program runs with administrator rights, which
// is when it can write to Program Files. It only asks; it never tries to obtain
// them (an unsigned program that starts an elevated step is treated as malware).
func isElevated() bool { return windows.GetCurrentProcessToken().IsElevated() }

func rootFor(everyone bool) registry.Key {
	if everyone {
		return registry.LOCAL_MACHINE
	}
	return registry.CURRENT_USER
}

func setValues(root registry.Key, values []registryValue) error {
	for _, v := range values {
		key, _, err := registry.CreateKey(root, v.Key, registry.SET_VALUE)
		if err != nil {
			return err
		}
		err = key.SetStringValue(v.Name, v.Value)
		_ = key.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// registerUninstall lists the program in Settings > Apps.
func registerUninstall(t target, version string) error {
	root := rootFor(t.Everyone)
	if err := setValues(root, uninstallValues(t, version)); err != nil {
		return err
	}
	key, err := registry.OpenKey(root, uninstallKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = key.Close() }()
	for name, value := range uninstallSwitches {
		if err := key.SetDWordValue(name, value); err != nil {
			return err
		}
	}
	return nil
}

// registerFiles makes .cuppa files open in the program. The person's own
// registration is always written; for everyone it is written for the computer.
func registerFiles(t target) error {
	values := associationValues(t.Program)
	err := setValues(registry.CURRENT_USER, values)
	if t.Everyone {
		err = errors.Join(err, setValues(registry.LOCAL_MACHINE, values))
	}
	notifyAssociationsChanged()
	return err
}

// notifyAssociationsChanged tells the shell that file types changed, so icons
// and "Open with" follow without signing out.
func notifyAssociationsChanged() {
	const assocChanged = 0x08000000
	shell := windows.NewLazySystemDLL("shell32.dll")
	_, _, _ = shell.NewProc("SHChangeNotify").Call(assocChanged, 0, 0, 0)
}

// removeRegistrations takes the program out of Settings > Apps and its file
// type out of the registry, for the person and, when the account may, for
// everyone. A key that is not there is not a problem; a file type that another
// program has taken over is left alone.
func removeRegistrations() error {
	var errs []error
	for _, root := range []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE} {
		errs = append(errs, deleteKey(root, uninstallKey))
		if ours(root) {
			for _, k := range associationKeys() {
				errs = append(errs, deleteKey(root, k))
			}
			errs = append(errs, deleteKey(root, classesKey+`\`+extension))
		}
	}
	notifyAssociationsChanged()
	return errors.Join(errs...)
}

// ours reports whether .cuppa is still registered to this program's type.
func ours(root registry.Key) bool {
	key, err := registry.OpenKey(root, classesKey+`\`+extension, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer func() { _ = key.Close() }()
	v, _, err := key.GetStringValue("")
	return err == nil && v == progID
}

func deleteKey(root registry.Key, path string) error {
	err := registry.DeleteKey(root, path)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}

// makeShortcut creates a Start menu shortcut with the Windows scripting host.
func makeShortcut(link, program string) error {
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShell(shortcutScript))
	cmd.Env = append(os.Environ(), "CUPPA_SHORTCUT="+link, "CUPPA_TARGET="+program, "CUPPA_FOLDER="+filepath.Dir(program))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	return cmd.Run()
}

// startInstalled starts the installed program with args, detached from this one.
func startInstalled(t target, args []string) error {
	cmd := exec.Command(t.Program, args...)
	cmd.Dir = t.Folder
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP}
	return cmd.Start()
}

// removeRunningProgram gets the running program out of the way. Windows will
// not delete a program that is running, but it does allow it to be renamed, so
// the program is moved into the temporary folder, which Windows empties by
// itself, and the folder it was in can then be removed.
func removeRunningProgram(path string) (leftover string, err error) {
	aside := filepath.Join(os.TempDir(), "cuppa-removed-"+strconv.Itoa(os.Getpid())+".exe")
	if err := os.Rename(path, aside); err == nil {
		return "", nil
	}
	inPlace := path + ".removed"
	if err := os.Rename(path, inPlace); err != nil {
		return "", err
	}
	return inPlace, nil
}
