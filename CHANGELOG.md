# Changelog

All notable changes to VibeShield. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versioning is [SemVer](https://semver.org/spec/v2.0.0.html).

## [2.0.0] — 2026-09-18

The release that turns the scanner into a **terminal workspace**. `vibeshield`
with no arguments now opens a searchable, keyboard-driven console over every
action, rule and agent recipe — and the same index is available to scripts and
AI agents through `vibeshield search`.

### Added

- **`--format sarif`.** SARIF 2.1.0 output, which `contracts/cli.md` has listed
  as a supported format since v1.0.0 while the binary answered "lands in v2.1".
  Findings now upload to GitHub code scanning via
  `github/codeql-action/upload-sarif` and appear in a repository's Security tab
  with the rule text, the one-line fix, and a `partialFingerprints` entry built
  from the contract's own dismiss hash so an alert survives reformatting.
  Severity maps to SARIF levels (`critical`/`high` → `error`, `medium` →
  `warning`, `low`/`info` → `note`); line-less findings omit the region rather
  than emitting the invalid `startLine: 0`.
- **`-v` / `--verbose`** on `scan` and `fix` — the global flag
  `contracts/cli.md` documented but the binary never accepted. Reports the
  resolved config, mode, languages, pack version and rule count, plus the walk
  result, on stderr.
- **`scripts/audit-contract.mjs`** — probes the built binary for every command,
  flag and format value the docs promise, and fails when one is missing. This
  is how `init` and `--format sarif` were found.
- **`vibeshield init`.** The setup command that `contracts/cli.md` has always
  specified but the binary never implemented. Detects the stack from manifests
  (language, ecosystem, framework), writes a `vibeshield.yml`, a pull-request
  gate workflow pinned to the release this binary came from, and a POSIX
  pre-commit hook, then runs the first scan. Conservative by design: it never
  overwrites a file without `--force`, never touches an existing git hook, and
  `--dry-run` prints the whole plan first. Flags: `--mode`, `--dry-run`,
  `--force`, `--no-hook`, `--no-workflow`, `--no-scan`, `--no-color`.
- **Interactive console (`vibeshield`, `vibeshield ui`).** One search box over
  the actions, 7 agent setup recipes and all 122 core rules. Arrow keys / `Tab`
  to move, `Enter` to open, `Esc` to clear then quit, `Ctrl+U` to reset.
  Selecting an action hands the terminal back and runs the real command, so the
  menu can never drift from the documented flags.
- **`vibeshield search [query]`.** The same ranking on stdout, with `--list`,
  `--rules`, `--agents`, `--limit` and `--format json` (`vibeshield.search/v1`)
  for scripting and agent tool-calls.
- **`vibeshield agents [name]`.** Per-agent setup recipes for Codex, Claude
  Code (CLI + Desktop), Google Antigravity, Cursor, Windsurf, GitHub Copilot and
  any `AGENTS.md` client. `--body` prints the pasteable rule block, `--markdown`
  emits the matrix as a table.
- **Fuzzy search engine** (`scanner/internal/cli`): separator folding
  (`vs-sec-017` ≡ `vs sec 017` ≡ `vssec017`), word-boundary ranking, keyword
  reinforcement and a subsequence fallback. Multi-token queries narrow rather
  than widen. Fully unit-tested, no dependencies.
- **First-class agent rule files, generated from one source of truth**:
  `AGENTS.md`, `.agents/rules/vibeshield.md` (Antigravity), `.cursor/rules/vibeshield.mdc`,
  `.cursorrules`, `.windsurf/rules/vibeshield.md`, `.windsurfrules`,
  `.github/copilot-instructions.md`. `scripts/sync-agent-rules.mjs --check`
  fails CI if any of them drifts from the Go constant.
- **`scripts/bump-version.mjs`** — explicit, auditable version bump that never
  touches fixture project versions.
- **`CHANGELOG.md`**, `docs/agents.md`.

### Changed

- **Version 2.0.0** across the binary, npm launcher, GitHub Action, Claude Code
  plugin, rule packs, contracts and site docs.
- `vibeshield` with no arguments: on a TTY it opens the console; when stdin is a
  pipe it still prints usage and exits `2`, so existing automation is unchanged.
- Rule packs bumped to `2.0.0` (rule IDs and semantics unchanged — this is a
  pack-version bump, not a rule rewrite).

### Fixed

- Rule-pack version drift between `rules/core/` and the embedded copy under
  `scanner/internal/rules/packs/core/`.

### Notes

- Zero new runtime dependencies. Raw-terminal handling is pure `syscall` per
  platform (`term_windows.go` / `term_linux.go` / `term_darwin.go`), so the
  scanner still builds as a single static binary with no CGO.
- Cross-compiles clean for linux/amd64, linux/arm64, darwin/amd64,
  darwin/arm64, windows/amd64 and windows/arm64.

## [1.0.0] — 2026-09-13

First public release: Go scanner with the embedded MIT core rule pack, GitHub
Action PR gate, pre-commit hook, npm launcher, Claude Code audit skill, and the
Astro site.
