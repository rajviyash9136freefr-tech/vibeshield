# VibeShield — repo conventions (READ FIRST)

You are working on VibeShield: a security & dependency auditor for AI-generated code.
The repository is the source of truth for behaviour. When a doc and the binary
disagree, the binary wins and the doc is a bug — `scripts/audit-contract.mjs`
exists to catch exactly that, and it fails CI when a documented command, flag or
format value is missing from the built scanner.

## Repo layout (monorepo)

```
scanner/   Go scanner binary (CLI, console, rules engine, VibePatch, doctor)
rules/     Versioned YAML rule packs (core pack = MIT, ships embedded in the binary)
action/    GitHub Action composite wrapper (action.yml + entrypoint script)
site/      Astro 5 marketing site & docs (GitHub Pages, base = /vibeshield/)
npm/       npm wrapper package (fetch-on-first-run launcher for the Go binary)
skill/     Claude Code plugin + the Agent Skills-format bug-hunting skill
contracts/ Cross-component JSON schemas + docs — the interface law
fixtures/  Golden test repos & seeded AI-PR corpus
docs/      Product docs source (agents.md, growth.md)
scripts/   Build/dev helpers: installers, sync-rules, sync-agent-rules,
           audit-contract, bump-version
```

There is no `api/` directory and no server component. Do not add one without a
deliberate decision: the whole privacy promise rests on there being nowhere for
findings to go.

## Distribution model (decided 2026-09-13)

VibeShield is **fully free and MIT-licensed end to end** — scanner, action, hooks,
console, all rule packs. No pricing tiers exist anywhere (site, docs, code) — do
not introduce any. The npm package `vibeshield` is a launcher that downloads and
checksum-verifies the release binary on first run; the real binary is Go,
published via GitHub Releases by `.github/workflows/release.yml` with asset names
the Action entrypoint and the npm launcher both resolve.

**Versioning:** the tool version and the rule-pack version are separate. The pack
carries `version:` in `rules/core/*.yaml`; a CLI-only release must leave it alone.
`node scripts/bump-version.mjs <from> <to>` does the tool bump and refuses to
guess — pass `--pack <v>` only when the rules themselves changed.

## Hard rules

1. **Contracts are law.** `contracts/finding/schema.json` defines the finding format.
   `contracts/cli.md` defines commands, flags, exit codes and the JSON report.
   Never fork either locally.
2. **No code leaves the machine.** Static analysis only. The scanner links in no
   HTTP client at all — keep it that way. `--online` is *reserved*: it prints a
   notice and the scan continues offline. Do not describe it as working.
3. **Tone:** blame patterns, never people or vendors. No fear-mongering. Numbers over
   adjectives. Banned words: "revolutionary", "game-changing", "military-grade", "effortless".
4. **Do not claim what the tool cannot do.** There is no dashboard, no cloud API,
   no OAuth, and no brew formula unless one has actually been published. Five rules
   are reserved and cannot fire — `vibeshield version` says so, and so must the docs.
5. **Windows-first dev environment:** build scripts must work in Git Bash on Windows.
   Go binary is pure Go (no CGO). Site is plain Node — no platform-specific postinstall.
   `.gitattributes` pins line endings; do not remove it.
6. **Go toolchain:** Go 1.24+. If `go` is not on PATH, look in
   `~/.workbuddy-ai/binaries/go/go/bin`. `gofmt -l .` only means something on an
   LF checkout — on a CRLF working tree it flags every file.
7. Every Go package gets `_test.go` files for exported behaviour. Rules engine and
   help/completion tables are table-driven.
8. Severity tokens: CRITICAL/HIGH/MEDIUM/LOW/INFO. UI colours ONLY from UIUX.md §2.1.
9. Rule IDs: `VS-PKG-###` (packages), `VS-SEC-###` (secrets/insecure patterns),
   `VS-LIC-###` (license), `VS-DEP-###` (dependency tree), `VS-INJ-###` (prompt-injection).
10. Commits: conventional commits (`feat(scanner): ...`). Never commit real secrets —
    the fixtures use obviously-fake keys (`sk-proj-FAKE...`).

## The four gates (run these before you claim a change is done)

```bash
cd scanner && go test ./... && go vet ./... && gofmt -l .
node scripts/sync-rules.mjs --check          # rules/core ↔ embedded pack
node scripts/sync-agent-rules.mjs --check    # generated agent rule files
node scripts/audit-contract.mjs scanner/vibeshield   # docs ↔ binary
```

## Adding a command or flag

`scanner/cmd/vibeshield/help.go` is the single source of truth for the first
three; `help_test.go` fails the build when they drift apart:

1. `docs()` — the help entry (name, group, summary, usage, body).
2. `completionFlags` in `completion.go` — the flags Tab should offer.
3. `scripts/audit-contract.mjs` — the `COMMANDS` list, for a new verb.
4. The docs: `README.md` and `site/src/content/docs/`.

## Definition of done for any component

- Builds/renders clean from a fresh checkout on this machine.
- Tests pass (`go test ./...` / `npm run build` in `site/`).
- Matches its contract in `contracts/`.
- Every documented surface exists in the binary (gate 4 above).
- No TODO left where user-visible copy should be.
