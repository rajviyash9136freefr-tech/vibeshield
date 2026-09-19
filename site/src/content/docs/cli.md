---
title: CLI reference
description: "vibeshield scan, init, version — flags, output formats (pretty/json/sarif/github), exit codes, and the offline-by-default promise."
order: 4
---

The `vibeshield` binary is a single static Go build — no runtime deps, boots in
under 50 ms, works offline. It's what the [Action](/docs/github-action) and the
[hook](/docs/pre-commit) run underneath, and it's usable standalone anywhere.

```bash
brew install vibeshield           # macOS/Linux
npx vibeshield scan .             # zero-install
go install github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@latest
```

## Commands

| Command | Meaning |
|---|---|
| `vibeshield scan [path]` | Scan a directory (full mode) or a diff (`--diff`/`--staged`). |
| `vibeshield init` | Detect frameworks, write `vibeshield.yml` + hook + workflow, run a first scan. |
| `vibeshield version` | Print version + embedded rule-pack versions. |

## Flags

```
--diff <ref|->     diff mode vs a git ref, or unified diff on stdin
--staged           scan git staged changes (pre-commit mode)
--format <fmt>     pretty | json | sarif | github   (default pretty on TTY)
--config <file>    config path (default vibeshield.yml if present)
--online           allow package-intel network lookups (default off)
--mode <mode>      off | warn | block-on-critical | block-on-high+
--max-cols <n>     output width cap (default 88)
--no-color         disable color (also: NO_COLOR env, non-TTY auto)
-v / --verbose
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Clean, or findings present in `warn`/`off` mode |
| 1 | Findings at/above the block threshold in a `block-*` mode |
| 2 | Config or usage error |

## JSON output

`--format json` emits the machine report every surface renders from:

```json
{
  "schema_version": 1,
  "tool": "vibeshield",
  "version": "2.0.1",
  "scan": { "mode": "diff", "ref": "HEAD~1", "files_scanned": 14, "duration_ms": 1234 },
  "summary": { "critical": 1, "high": 1, "medium": 0, "low": 0, "info": 0, "clean_files": 11 },
  "dependencies": [
    { "name": "fast-parse-utils-v3", "version": "2.1.4", "ecosystem": "npm",
      "age_days": 9, "maintainers": 1, "risk": "critical",
      "why": "registered 9 days ago; 1 maintainer; post-install fetches remote binary" }
  ],
  "findings": [ /* contracts/finding/schema.json objects */ ]
}
```

`--format sarif` targets GitHub code-scanning uploads; `--format github` emits
`:error:`/`:warning:` workflow annotations inline in the diff.

## Privacy

Static analysis only. `--online` is opt-in and sends **package names**, never
code: the scanner's default posture is no network. (The in-agent
[bug-hunting skill](/docs/skill) is offline by design too.)
