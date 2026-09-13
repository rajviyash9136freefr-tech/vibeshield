# VibeShield — repo conventions (READ FIRST)

You are working on VibeShield: a security & dependency auditor for AI-generated code.
Specs live in `PRD.md` (features, copy, pricing) and `UIUX.md` (design, pages, SEO).
**Both specs are the source of truth.** When code and spec disagree, follow the spec
or flag it — never silently improvise product copy or numbers.

## Repo layout (monorepo)

```
scanner/   Go scanner binary (CLI + pre-commit engine). Module: github.com/rajviyash9136freefr-tech/vibeshield/scanner
rules/     Versioned YAML rule packs (core pack = MIT, ships with the binary)
action/    GitHub Action composite wrapper (action.yml + entrypoint script)
api/       Fastify package-intel + findings API (Node 22+ / TypeScript)
site/      Astro marketing site (vibeshield.dev) + Tailwind
npm/       npm wrapper package (fetch-on-first-run launcher for the Go binary)
contracts/ Cross-component JSON schemas + docs — the interface law
fixtures/  Golden test repos & seeded AI-PR corpus
docs/      Product docs source (quickstart, GitHub Action page)
scripts/   Build/dev helper scripts (sync-rules.mjs keeps embedded packs honest)
```

## Distribution model (decided 2026-09-13)

VibeShield is **fully free and MIT-licensed end to end** — scanner, action, hooks,
all rule packs. No pricing tiers exist anywhere (site, docs, code) — do not
introduce any. The npm package `vibeshield` is a thin launcher (no install-time
network); the real binary is Go, published via GitHub Releases by
`.github/workflows/release.yml` with asset names the Action entrypoint downloads.

## Hard rules

1. **Contracts are law.** `contracts/finding/schema.json` defines the finding format.
   Scanner emits it; API stores it; site/tests validate against it. Never fork it locally.
2. **No code leaves the machine** (CLI/hook): static analysis only, network calls are
   opt-in (`--online` for package-intel lookups, degrade gracefully offline).
3. **Tone:** blame patterns, never people or vendors. No fear-mongering. Numbers over
   adjectives. Banned words: "revolutionary", "game-changing", "military-grade", "effortless".
4. **Windows-first dev environment:** build scripts must work in Git Bash on Windows.
   Go binary is pure Go (no CGO). Site/API are plain Node — no platform-specific postinstall.
5. **Go toolchain** is portable at `W:\tools\go\bin\go.exe` (add to PATH per-command:
   `export PATH="/w/tools/go/bin:$PATH"`). Go 1.27.1.
6. Every Go package gets `_test.go` files for exported behavior. Rules engine is table-driven.
7. Severity tokens: CRITICAL/HIGH/MEDIUM/LOW/INFO. UI colors ONLY from UIUX.md §2.1.
8. Rule IDs: `VS-PKG-###` (packages), `VS-SEC-###` (secrets/insecure patterns),
   `VS-LIC-###` (license), `VS-DEP-###` (dependency tree), `VS-INJ-###` (prompt-injection).
9. Commits: conventional commits (`feat(scanner): ...`). Never commit real secrets —
   the fixtures use obviously-fake keys (`sk-proj-FAKE...`).

## Definition of done for any component

- Builds/renders clean from a fresh checkout on this machine.
- Tests pass (`go test ./...` / `npm test`).
- Matches its contract in `contracts/`.
- No TODO left where user-visible copy should be.
