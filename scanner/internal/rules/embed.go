// Embedded core rule pack. The canonical copies live in rules/core/ (the
// versioned, diffable source of truth); this file embeds them so the binary
// ships self-contained and works fully offline. scripts/sync-rules.mjs keeps
// packs/core in sync with rules/core before builds.
package rules

import (
	"embed"
	"io/fs"
	"sort"
)

//go:embed packs/core/*.yaml
var coreFS embed.FS

// CorePackFiles lists the embedded core pack filenames (for `version`).
func CorePackFiles() []string {
	var names []string
	fs.WalkDir(coreFS, "packs/core", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			names = append(names, path)
		}
		return nil
	})
	sort.Strings(names)
	return names
}

// LoadCore parses and merges every embedded core pack document. The merge
// stamps each rule with its owning pack, and duplicate rule IDs across
// documents are an error (contracts/rulepack.md: ids unique across packs).
func LoadCore() (*Pack, error) {
	merged := &Pack{}
	seen := map[string]bool{}
	for _, name := range CorePackFiles() {
		data, err := coreFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		p, errs := ParsePackBytes(data, name, LoadOptions{})
		if len(errs) > 0 {
			return nil, errs[0]
		}
		if merged.ID == "" {
			merged.Schema, merged.ID, merged.Version, merged.License = p.Schema, p.ID, p.Version, p.License
		}
		for _, r := range p.Rules {
			if seen[r.ID] {
				return nil, errDuplicateID(r.ID)
			}
			seen[r.ID] = true
			merged.Rules = append(merged.Rules, r)
		}
	}
	return merged, nil
}

type dupErr struct{ id string }

func (d dupErr) Error() string { return "duplicate rule id " + d.id }

func errDuplicateID(id string) error { return dupErr{id} }
