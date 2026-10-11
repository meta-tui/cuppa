package install

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// place copies the program at src to dst, creating dst's folder. It writes to a
// temporary name first, so a copy that is interrupted never leaves half a
// program where a shortcut points. An existing program at dst is replaced,
// which fails, saying so, if it is running.
func place(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	partial := dst + ".part"
	out, err := os.OpenFile(partial, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(partial)
		return err
	}
	if _, statErr := os.Stat(dst); statErr == nil {
		if err := os.Remove(dst); err != nil {
			_ = os.Remove(partial)
			return fmt.Errorf("the copy already there cannot be replaced, probably because it is running: %w", err)
		}
	}
	if err := os.Rename(partial, dst); err != nil {
		_ = os.Remove(partial)
		return err
	}
	return nil
}

// removeAfterExit deletes the program a move left behind, once the copy that
// started the move has exited. Windows will not delete a running program, so
// the first attempts fail and are repeated until one works or the attempts run
// out. A program that is already gone counts as done.
func removeAfterExit(path string, attempts int, interval time.Duration, remove func(string) error, sleep func(time.Duration)) bool {
	for range attempts {
		err := remove(path)
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			return true
		}
		sleep(interval)
	}
	return false
}

// CleanUpMovedProgram removes the copy a move left behind, about ten seconds
// of patience. It never removes the program that is running it.
func CleanUpMovedProgram(path string) {
	if path == "" {
		return
	}
	if exe, err := os.Executable(); err == nil && samePlace(goos, exe, path) {
		return
	}
	removeAfterExit(path, 50, 200*time.Millisecond, os.Remove, time.Sleep)
}
