// Package vscodepm — syncat.go: explicit-path siblings of the
// default Sync / ListEntries entry points.
//
// Default Sync / ListEntries always resolve the projects.json
// location through ProjectsJSONPath (XDG / APPDATA / portable-VS-Code
// discovery). The variants below skip that resolver and operate on
// any caller-supplied absolute path. Used by the
// `gitmap vscode-pm-sync --projects-json <path>` flag and by tests
// that need to point at a fixture file inside t.TempDir().
//
// Same merge semantics, same atomic-rename writer — these helpers
// only differ in WHERE the file lives, never in WHAT they do to it.
package vscodepm

// SyncAt is the explicit-path sibling of Sync. Used by callers that
// have already resolved (or been given) a projects.json path — e.g.
// `gitmap vscode-pm-sync --projects-json <path>` — and don't want
// the path resolver to second-guess them. Same merge / atomic-write
// semantics as Sync.
func SyncAt(path string, pairs []Pair) (SyncSummary, error) {
	existing, err := readEntries(path)
	if err != nil {
		return SyncSummary{}, err
	}

	merged, summary := mergePairs(existing, pairs)

	if err := writeEntriesAtomic(path, merged); err != nil {
		return summary, err
	}

	summary.Total = len(merged)

	return summary, nil
}

// ListEntriesAt is the explicit-path sibling of ListEntries. Used by
// callers that have already resolved (or been given) a projects.json
// path — e.g. `gitmap vscode-pm-sync --projects-json <path>` — and
// don't want the resolver to second-guess them.
func ListEntriesAt(path string) ([]Entry, error) {
	return readEntries(path)
}
