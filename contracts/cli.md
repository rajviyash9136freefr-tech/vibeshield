# VibeShield CLI contract (v1)

Binary name: `vibeshield` (Go, single static binary, pure Go — no CGO).

## Commands

```
vibeshield scan [path]     Scan a directory (full mode) or diff (diff mode)
vibeshield fix [path]      VibePatch: preview + apply mechanical autofixes (opt-in, gated)
vibeshield init [path]     Detect frameworks, write vibeshield.yml + hook + workflow, run first scan
vibeshield search [query]  Search the rule packs and the console catalog
vibeshield agents [name]   Print per-agent setup recipes (Codex, Claude Code, Antigravity, …)
vibeshield ui              Open the interactive console (same as a bare `vibeshield`)
vibeshield version         Print version + embedded rule-pack versions
```

## Init flags

```
--mode <mode>      Initial gate mode written to vibeshield.yml (default: warn)
--dry-run          Show what would be written; change nothing
--force            Overwrite files that already exist
--no-hook          Skip the git pre-commit hook
--no-workflow      Skip the GitHub Action workflow
--no-scan          Skip the first scan
--no-color         Disable color
```

Init law: it writes at most three files — `vibeshield.yml`,
`.github/workflows/vibeshield.yml`, `.git/hooks/pre-commit` — and never
overwrites any of them (least of all a git hook) without `--force`. With no
git repository present it skips the hook and still writes the other two. The
workflow pins the Action to the release the running binary came from, not to a
moving major tag.

## Fix flags (VibePatch)

```
--dry-run              Preview the −/+ diff; change nothing
--yes                  Skip the per-file prompt (coding-agent / CI mode). Every
                       applied patch is still appended to vibeshield-fixes.log
                       (JSONL: time, rule_id, file, line, before, after, mode)
--report <file>        Reuse a `scan --format json` file instead of rescanning
--config / --rules / --no-color   shared with scan
```

Gate law: `fix` prompts `[y/N/a/s]` per file; a non-TTY stdin must pass
`--dry-run` or `--yes` (it never guesses an answer). Rules without an
`autofix` (contracts/rulepack.md) are previewed as suggestions only.
`hardcoded-secret` and `prompt-injection` findings are never patched
mechanically — a key needs rotation, and an agent editing its own
instruction file is the injection we came to catch.

## Global flags

```
--diff <ref|->         Diff mode: scan changes vs git ref, or read unified diff from stdin (-)
--staged               Pre-commit mode: scan git staged changes (implies --diff on index)
--format <fmt>         pretty (default, TTY) | json | sarif | github  (github = :error/:warning annotations)
--config <file>        Config path (default: vibeshield.yml if present)
--online               Allow network calls for package-intel (default: off; offline heuristics used otherwise)
--mode <mode>          Override config mode: off | warn | block-on-critical | block-on-high+
--max-cols <n>         Output width cap (default 88 in TTY)
--no-color             Disable color (also: NO_COLOR env, non-TTY auto)
-v, --verbose          Explain what was scanned, on stderr (scan, fix)
```

`--verbose` is a subcommand flag, not a top-level one: bare `vibeshield -v` is
the version alias (see Commands), while `vibeshield scan -v` is verbose. The
scopes do not overlap.

## Exit codes (law — documented in --help)

| Code | Meaning |
|---|---|
| 0 | Clean, or findings present in warn/off mode |
| 1 | Findings present at-or-above block threshold (block-on-critical / block-on-high+) |
| 2 | Config / usage error |

## Config file: vibeshield.yml (schema)

```yaml
mode: warn                 # off | warn | block-on-critical | block-on-high+
languages: [typescript, python]   # optional filter; empty = all
ignore:                    # auditable ignores
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"   # reason is REQUIRED, echoed in reports
thresholds:
  new_dependency_max_age_days: 30
notifications:
  slack: ${SLACK_WEBHOOK}   # env interpolation only; never literal secrets
```

### Config resolution

An explicit `--config <file>` always wins. Otherwise the scanner looks for
`vibeshield.yml` in the **scanned directory** first, then falls back to the
working directory, then to defaults. Scanning another project therefore honours
that project's configuration instead of silently ignoring it — the bug this
rule exists to prevent.

### notifications.slack

Must be an environment reference (`${SLACK_WEBHOOK}` or `$SLACK_WEBHOOK`).
Anything else — including a webhook URL, or a malformed reference like `${A` —
is a config error (exit 2). A webhook URL committed to a config file is a
leaked credential: the path *is* the whole secret, so the mistake must not be
mergeable quietly. Delivery itself is not wired up in this build; the block is
validated so a config can be written once and stay correct.

## JSON output (stdout, `--format json`)

```json
{
  "schema_version": 1,
  "tool": "vibeshield",
  "version": "2.0.1",
  "scan": { "mode": "diff", "ref": "HEAD~1", "files_scanned": 14, "duration_ms": 1234 },
  "summary": { "critical": 1, "high": 1, "medium": 0, "low": 0, "info": 0, "clean_files": 11 },
  "dependencies": [
    { "name": "fast-parse-utils-v3", "version": "2.1.4", "ecosystem": "npm",
      "age_days": 9, "maintainers": 1, "risk": "critical", "why": "registered 9 days ago; 1 maintainer; post-install script fetches remote binary" }
  ],
  "findings": [ /* contracts/finding/schema.json objects */ ]
}
```

## Output format (pretty, TTY) — the VibeCheck format

Widths ≤ 88 cols. Severity dot colors: critical #F87171, high #FB923C, medium #FBBF24,
low #60A5FA, info #71717A. Non-TTY: plain text, no emoji color codes, emojis kept.

```
$ vibeshield scan --staged

  Scanning 14 changed files (diff mode)… done in 1.2s

  🔴 CRITICAL  VS-PKG-001  hallucinated-package
     fast-parse-utils-v3@2.1.4 — registered 9 days ago, 1 maintainer,
     post-install script fetches remote binary.
     → Fix: replace with node:util (12-line change in src/parse.ts)

  🟠 HIGH      VS-SEC-017  hardcoded-secret
     OPENAI_API_KEY echoed in src/lib/agent.ts:41 — likely pasted
     from an AI chat response.
     → Fix: move to env, rotate the key now

  ✓ 11 files clean · 2 findings · 1 dependency added (risky)
```

## Attribution (AI-origin) signals, in priority order

1. `confirmed`: commit author matches `*[bot]` / known agent logins (copilot-swe-agent,
   cursor-bot, …) OR `Co-authored-by: .*[bot] <` trailer OR `Generated-by`/`Co-authored-by:
   Claude` style trailers OR file contains an AI-generated marker comment on the hunk.
2. `likely`: hunk style heuristics (see scanner/README) score ≥ threshold.
3. `unknown`: otherwise.

## Performance budget (PRD 5.2 / UIUX 8)

- Pre-commit staged scan: < 1.5 s on typical diffs.
- Action PR scan: < 10 s median.
- Binary boot: < 50 ms.
