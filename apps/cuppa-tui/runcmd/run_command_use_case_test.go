package runcmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/runhost/local"
)

func designFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.cuppa")
	if _, err := disk.Save(path, design.NewDocument("App", 40, 10)); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCmd(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := Run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestSetStoresTheBlockInTheDesignAndShowReadsItBack(t *testing.T) {
	path := designFile(t)
	code, _, errs := runCmd("set", path,
		"--build", `go build -o "C:/my dir/app.exe" ./cmd/app`,
		"--run", `"C:/my dir/app.exe" --demo`,
		"--env", "APPDATA=C:/scratch", "--env", "HOME=C:/home",
		"--build-dir", "..", "--dir", "C:/scratch/work", "--size", "120x36", "--on-save")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	doc, err := disk.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	r := doc.Run
	if len(r.Build) != 5 || r.Build[3] != "C:/my dir/app.exe" || len(r.Run) != 2 || r.Run[1] != "--demo" ||
		r.Env["APPDATA"] != "C:/scratch" || r.BuildDir != ".." || r.Dir != "C:/scratch/work" ||
		r.Width != 120 || r.Height != 36 || !r.OnSave {
		t.Fatalf("got %+v", r)
	}
	code, out, _ := runCmd("show", path)
	if code != 0 || !strings.Contains(out, "demo") || !strings.Contains(out, "APPDATA=C:/scratch") {
		t.Fatalf("show: %d %q", code, out)
	}
}

func TestSetChangesOnlyWhatIsGiven(t *testing.T) {
	path := designFile(t)
	if code, _, e := runCmd("set", path, "--run", "app", "--env", "A=1"); code != 0 {
		t.Fatal(e)
	}
	if code, _, e := runCmd("set", path, "--env", "B=2", "--on-save"); code != 0 {
		t.Fatal(e)
	}
	doc, _ := disk.Load(path)
	if len(doc.Run.Run) != 1 || doc.Run.Env["A"] != "1" || doc.Run.Env["B"] != "2" || !doc.Run.OnSave {
		t.Fatalf("got %+v", doc.Run)
	}
}

func TestLocalWritesTheOverrideAndLeavesTheDesignAlone(t *testing.T) {
	path := designFile(t)
	if code, _, e := runCmd("set", path, "--run", "app", "--dir", "x"); code != 0 {
		t.Fatal(e)
	}
	if code, _, e := runCmd("set", path, "--local", "--dir", "C:/scratch", "--env", "APPDATA=C:/s"); code != 0 {
		t.Fatal(e)
	}
	doc, _ := disk.Load(path)
	if doc.Run.Dir != "x" || len(doc.Run.Env) != 0 {
		t.Fatalf("the design changed: %+v", doc.Run)
	}
	merged, err := local.Load(path, doc.Run)
	if err != nil || merged.Dir != "C:/scratch" || merged.Env["APPDATA"] != "C:/s" {
		t.Fatalf("%+v %v", merged, err)
	}
}

func TestARunWithoutAProgramIsRefused(t *testing.T) {
	path := designFile(t)
	if code, _, _ := runCmd("set", path, "--build", "go build"); code == 0 {
		t.Fatal("a block with nothing to run was stored")
	}
	if code, _, _ := runCmd("set", path, "--size", "big", "--run", "app"); code == 0 {
		t.Fatal("a bad size was accepted")
	}
}

func TestClearRemovesTheSettings(t *testing.T) {
	path := designFile(t)
	if code, _, e := runCmd("set", path, "--run", "app"); code != 0 {
		t.Fatal(e)
	}
	if code, _, e := runCmd("set", path, "--clear"); code != 0 {
		t.Fatal(e)
	}
	doc, _ := disk.Load(path)
	if doc.Run.IsSet() {
		t.Fatalf("still set: %+v", doc.Run)
	}
}
