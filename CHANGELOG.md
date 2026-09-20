# Changelog

All notable changes to VibeShield. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versioning is [SemVer](https://semver.org/spec/v2.0.0.html).

## [3.0.1] — 2026-09-20

### Added
- **Interactive Terminal UI:** Bubble Tea interactive terminal application (`vibeshield`) featuring live stack detection, slash commands (`/scan`, `/diff`, `/fix`, `/help`, `/quit`), streaming scanner progress, navigable findings table, code context viewer, and mechanical VibePatch preview.
- **Asynchronous Streaming Scan Engine:** `scan.Stream` for non-blocking real-time event updates across files and discovered findings.
- **Comprehensive QA Test Suite:** Added `test/` suite with multi-stack vulnerable fixtures (`vulnerable-node-app`, `vulnerable-react-app`, `vulnerable-supabase-app`, `vulnerable-python-app`, `clean-app`) and edge-case suites (unicode paths, binary files, monorepos).
- Added `test/COMMANDS.md` full inventory checklist and `test/BUGS.md` QA bug tracker.

### Fixed
- Fixed `scandiff.Diff` line membership method invocation in interactive diff mode.
- Synchronized local install binary path resolution in PowerShell scripts.
- Tidied indirect Go module checksums for Charm Bubbles.

## [3.0.0] — 2026-09-19

The release that makes VibeShield **usable by someone who has never seen it**.
v1 was a scanner, v2 was a workspace — but a workspace whose manual you had to
read first. v3 is the version you can hand to a colleague: every command
documents itself, a typo names the command you meant, `doctor` tells you what
is wired up and what is not, and Tab completes the flags.

The rule pack is **unchanged at `2.0.0`** — this is a tool release, not a rule
rewrite. `vibeshield version` reports `3.0.0` for the binary and `core 2.0.0`
for the pack, on purpose.

### Added

- **`vibeshield doctor [path]`.** A read-only health check for the binary, the
  config, the git pre-commit hook, the PR-gate workflow and the agent rule
  files. Every gap is printed with the exact command that fixes it. `--format
  json` emits `vibeshield.doctor/v1`; the exit code is `1` when anything needs
  fixing, so CI can assert the gate is really wired up. This is the first thing
  to run when a scan behaves unexpectedly.
- **`vibeshield completion <bash|zsh|fish|powershell>`.** Generated completion
  scripts. Every verb and flag comes from the same tables the parser uses, and
  a test fails the build if the help text and the completion tables drift.
- **`vibeshield rules [id]`.** The rule packs without the actions and agent
  recipes that `search` also indexes — `search --rules --limit 0` under a name
  people actually reach for. `vibeshield rules VS-SEC-017` reads one rule in
  full; `vibeshield rules` lists them all, grouped by category.
- **Per-command help.** `vibeshield scan --help`, `vibeshield help scan` and
  `vibeshield scan -h` now print that command's own flags and examples, and
  exit `0`. Previously every subcommand printed the global manual, so
  `scan --help` never mentioned `--diff`.
- **"Did you mean" suggestions.** `vibeshield scna` now answers with
  `Did you mean \`vibeshield scan\`?` instead of the whole manual. Similarity is
  Damerau-Levenshtein based, so transposed letters count as one edit, and the
  0.6 floor means an unrelated word gets silence rather than a confident wrong
  answer.
- **`-V` as a version alias**, alongside the existing `-v` and `--version`.
- `.gitattributes` normalising line endings. Without it a Windows checkout
  rewrites every file to CRLF, which made `gofmt -l .` report every Go file as
  unformatted and put CRLF shebangs on the shell scripts.
- **`vibeshield.yml`, and the repository now scans itself in CI.** The
  "VibeShield scans itself" step in `.github/workflows/ci.yml` runs the scanner
  against this repository with `mode: block-on-critical`. Getting it to pass
  took a real ignore list, and the reason is instructive: a scanner pointed at
  its own source always finds itself. The rule packs contain every pattern they
  detect, the golden fixtures are vulnerable on purpose, and the docs must
  display the install one-liner that `VS-SEC-027` exists to flag. Every
  exception in `vibeshield.yml` names its files and states why, and nothing is
  excluded because a finding was inconvenient — the first run went from 248
  findings to 0 by fixing two real bugs and documenting the structural ones.
- `packaging/homebrew/README.md` and `scripts/gen-homebrew-formula.mjs`, which
  turns a published release's own `sha256sums.txt` into a ready-to-submit
  formula. The generated `.rb` is deliberately not committed: a formula with
  placeholder checksums is worse than no formula.
- `.github/workflows/publish-npm.yml`, publishing the npm launcher with OIDC
  trusted publishing (no stored token) and `--provenance`, and refusing to
  publish when `package.json` or `lib/run.js` disagree with the release tag.

### Changed

- **Top-level help rewritten around examples.** It opens with four commands to
  run first, groups the verbs into Everyday / Explore / Meta, and states the
  exit-code contract — because that is the whole contract for a gate.
- `--list` on `search` is no longer truncated by the default `--limit 20`. The
  README documents `vibeshield search --rules --list` as "every shipped rule"
  and it printed 20 entries. An explicit `--limit` still wins.
- The console catalog covers `doctor`, `rules` and `completion`.
- **Version 3.0.0** across the binary, npm launcher, GitHub Action, Claude Code
  plugin, contracts, docs and site.

### Fixed

- **The one-line installers were pinned to `v1.0.0`.** `scripts/install.sh` and
  `scripts/install.ps1` both defaulted to `v1.0.0`, and neither was listed in
  `scripts/bump-version.mjs` — so every release since the first bumped 26
  version touchpoints and silently skipped the two files that decide what the
  README's headline install command actually downloads. Anyone running that
  one-liner got a release two major versions old, or a 404 if the old assets
  had been pruned. Both now default to `v3.0.0` and both are in the bump list,
  so this cannot drift again.
- **The installers advised installing at a floating tag.** The `go install`
  fallbacks in `install.sh`, `install.ps1` and the npm launcher all ended in
  `@latest`, which is the exact pattern this project's own `VS-DEP-011` rule
  flags: a floating tag resolves to whatever is published at run time. They now
  pin to the release being installed. The npm fallback in `install.ps1` was
  also unpinned (`npm install -g vibeshield`) and now installs the matching
  version.
- **`publish-npm.yml` upgraded the global npm CLI from a floating tag.** A
  security tool's release workflow should not do the thing its own
  `VS-DEP-004` rule calls a supply-chain window. Node 24 bundles npm 11.x,
  which supports OIDC trusted publishing, so the upgrade is gone entirely and
  replaced with an assertion that fails loudly if a future Node downgrade makes
  trusted publishing unavailable.
- **`--rules` swallowed the following flag in `search`.** The positional-arg
  rewriter used one global table of value-taking flags, so in `search` — where
  `--rules` is a boolean switch — it consumed the next argument as if it were a
  directory name. `vibeshield search --rules --list` and
  `vibeshield rules VS-SEC-017 --no-color` both failed with a flag parse error.
  The table is now per-subcommand, and `TestFlagTakesValueCoversEveryValueFlag`
  covers the boolean/value split in both directions.
- `doctor --config <missing-file>` is now a warning rather than a silent
  fall-through to the defaults.
- `initcmd`'s exported `FindGitDir` replaces the unexported `findGitDir`, so
  `doctor` and `init` cannot disagree about where the git directory is.
- The `actionRef` test asserts against `defaultActionTag` instead of a literal
  version, so a release bump no longer has to edit a test.
- The Claude Code plugin manifest pointed at `github.com/vibeshield/vibeshield`
  and a `vibeshield.dev` homepage that does not exist. Both now point at the
  real repository and site.

### Notes

- Zero new runtime dependencies. Still a single static binary, no CGO.
- Cross-compiles clean for linux/amd64, linux/arm64, darwin/amd64,
  darwin/arm64, windows/amd64 and windows/arm64.

## [2.0.1] — 2026-09-19

Patch release: three documentation-vs-reality gaps found by auditing the docs
against the binary. No new features, no rule changes — the core pack stays at
`2.0.0`.

### Fixed

- **`notifications.slack` now enforces the rule the docs already claimed.**
  `contracts/cli.md` and the site config page both say a literal webhook URL is
  a config error (exit 2), because a webhook path *is* the credential. The
  field was not in the config struct at all, so `Load` silently ignored it and
  a committed webhook scanned clean. Now validated: `${SLACK_WEBHOOK}` and
  `$SLACK_WEBHOOK` are accepted; anything else — including the malformed
  `${A` — exits 2. Delivery is still not wired up; the block is validated so a
  config can be written once and stay correct.
- **`scan <path>` reads `<path>/vibeshield.yml`.** The config was resolved
  relative to the working directory, so `vibeshield scan ../other-project`
  silently used the wrong project's configuration, or none. An explicit
  `--config` still wins, and when the scan path is `.` nothing changes.
- **`vibeshield version` no longer overstates the rule count.** It reported
  "122 rules" when the engine can evaluate 117: five `structural` rules load
  and validate so packs stay portable, but the matcher skips them. It now
  reports `117 active · 5 reserved`, and the console marks a reserved rule as
  reserved in its detail pane instead of letting you search for a rule that
  cannot fire.

### Added

- `scripts/audit-contract.mjs` now reports rule coverage, so an inert-rule
  count is visible in CI rather than buried in the pack.
- `scripts/bump-version.mjs` takes `<from> <to>` and an optional
  `--pack <version>`, so a tool-only release can no longer accidentally bump
  the rule-pack version.

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
