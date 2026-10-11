package gosource

import (
	"fmt"
	"os"
	"path/filepath"
)

// Write puts the project in dir, creating it if needed. It refuses a folder
// that already holds one of the project's files, so it never overwrites work.
func Write(dir string, p Project) error {
	for name := range p.Files {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); err == nil {
			return fmt.Errorf("%s already has a %s; choose an empty folder", dir, name)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, content := range p.Files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
