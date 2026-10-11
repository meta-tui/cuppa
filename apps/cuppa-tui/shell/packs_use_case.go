package shell

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/packsdialog"
	"github.com/meta-ui-labs/cuppa/libs/catalog/bundled"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/catalog/packstate"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// restorePacks brings back which packs the user switched off and starts
// remembering changes. RestoreLayout calls it, so tests never touch settings.
func (m *Model) restorePacks() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	m.packs = packstate.Open(filepath.Join(dir, "cuppa", "packs.json"))
	m.refreshPalette()
}

// LoadUserPacks adds the packs bundled with Cuppa and the ones installed in the
// user's packs folder to the catalog. Front ends call it once at start; ReportPackProblems then tells the
// user about any file that could not be used. Tests do not call it, so they
// never read the user's folder.
func (m *Model) LoadUserPacks() {
	m.userPacks = cupp.UserDir()
	m.reloadPacks()
}

// reloadPacks rebuilds the catalog from the built-ins and the packs folder,
// e.g. after a pack was added, removed or created.
func (m *Model) reloadPacks() {
	packs, problems := bundled.Packs() // shipped inside Cuppa, listed first
	if m.userPacks != "" {
		installed, more := cupp.LoadDir(m.userPacks)
		packs, problems = append(packs, installed...), append(problems, more...)
	}
	reg, more := cupp.Extend(m.base, packs)
	m.packProblems = append(problems, more...)
	m.cat.Set(cupp.Adopt(reg, m.embedded))
	m.refreshPalette()
}

// adoptEmbedded makes the components a just-loaded design carries with it
// available, for a design opened on a computer without their packs.
func (m *Model) adoptEmbedded(doc design.Document) {
	m.embedded = doc.Embedded
	m.reloadPacks()
}

// refreshPalette lists only the components of enabled packs. The full catalog
// stays behind the editor, so a design that uses a disabled pack still renders.
func (m *Model) refreshPalette() {
	m.pal.SetCatalog(m.cat.Registry().Only(m.packs.Enabled))
}

// ReportPackProblems tells the user which installed packs could not be used.
func (m *Model) ReportPackProblems() {
	if len(m.packProblems) == 0 {
		return
	}
	var lines []string
	for _, p := range m.packProblems {
		lines = append(lines, "• "+p.Error())
	}
	m.packProblems = nil
	m.flow.Notice("Some packs could not be loaded", strings.Join(lines, "\n"))
}

// openPacks shows the Packs dialog; each switch applies at once.
func (m *Model) openPacks() {
	var entries []packsdialog.Entry
	for _, p := range m.cat.Packs() {
		entries = append(entries, packsdialog.Entry{Pack: p, Count: len(m.cat.ByFamily(p.ID))})
	}
	toggle := func(id definition.Family) {
		m.packs.SetEnabled(id, !m.packs.Enabled(id))
		m.refreshPalette()
	}
	m.flow.Show(packsdialog.New(entries, m.packs.Enabled, toggle), m.afterPacksDialog)
}
