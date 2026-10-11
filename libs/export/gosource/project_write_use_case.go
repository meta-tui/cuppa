package gosource

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// keepsOwn are the files of a project a person edits after the first export
// (the module's requirements, the notes): they are written when they are
// missing and left alone afterwards.
var keepsOwn = map[string]bool{"go.mod": true, "README.md": true}

func isGenerated(data []byte) bool { return strings.HasPrefix(string(data), GeneratedHeader) }

// ReplaceProject writes a project made of folders (the slice layout) into dir,
// again and again: a file Cuppa generated earlier is replaced, the generated
// files of a page that is gone are removed, and a file Cuppa did not generate
// is never touched; the call refuses the whole folder when one is in the way.
// go.mod and README.md are written once and then belong to the person.
func ReplaceProject(dir string, p Project) error {
	for name := range p.Files {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil || keepsOwn[name] {
			continue
		}
		if !isGenerated(data) {
			return fmt.Errorf("%s has a %s that Cuppa did not generate; an application needs a folder of its own", dir, name)
		}
	}
	for name, content := range p.Files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if _, err := os.Stat(path); err == nil && keepsOwn[name] {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	for _, stale := range staleFiles(dir, p) {
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(stale))); err != nil {
			return err
		}
	}
	return nil
}

// staleFiles are the generated Go files under dir that the project no longer
// has, as slash paths.
func staleFiles(dir string, p Project) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if _, kept := p.Files[rel]; kept {
			return nil
		}
		if data, err := os.ReadFile(path); err == nil && isGenerated(data) {
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

// CheckProject says how dir differs from what the project would write: a file
// that is missing or has other content, and a generated file the project no
// longer has. It writes nothing; an empty result means the folder is current.
// Line endings are not compared, and go.mod and README.md are not checked once
// they exist.
func CheckProject(dir string, p Project) []string {
	var out []string
	same := func(a, b string) bool {
		return strings.ReplaceAll(a, "\r\n", "\n") == strings.ReplaceAll(b, "\r\n", "\n")
	}
	names := make([]string, 0, len(p.Files))
	for name := range p.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		switch {
		case err != nil:
			out = append(out, name+": missing")
		case keepsOwn[name]:
		case !same(string(data), p.Files[name]):
			out = append(out, name+": differs from what the design exports")
		}
	}
	for _, stale := range staleFiles(dir, p) {
		out = append(out, stale+": generated for a page that is no longer designed")
	}
	return out
}
