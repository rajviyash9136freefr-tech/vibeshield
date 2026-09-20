---
title: CLI reference
description: "Every command and flag: scan, fix, init, doctor, search, rules, agents, completion. Output formats, exit codes, and the offline promise."
order: 4
---

The `vibeshield` binary is a single static Go build — no runtime deps, boots in
under 50 ms, works offline. It is what the [Action](/docs/github-action) and the
[hook](/docs/pre-commit) run underneath, and it is usable standalone anywhere.

```bash
vibeshield --version      # vibeshield 3.0.0
vibeshield help           # every command
vibeshield help scan      # one command, in full
```

Every command also takes `-h` / `--help`.

## Commands

| Command | What it does |
|:---|:---|
| `vibeshield init [path]` | Set a project up: config, PR gate, pre-commit hook, first scan |
| `vibeshield scan [path]` | Audit a directory, a git diff, or the staged changes |
| `vibeshield fix [path]` | Preview and apply the mechanical fixes (VibePatch) |
| `vibeshield doctor [path]` | Report what is wired up and what is not — changes nothing |
| `vibeshield search [query]` | Search every rule, action and agent recipe |
| `vibeshield rules [id]` | List the rule packs, or read one rule in full |
| `vibeshield agents [name]` | Per-agent setup recipes |
| `vibeshield completion <shell>` | Print a completion script |
| `vibeshield version` | Version, rule packs, engine coverage |
| `vibeshield help [command]` | Help |
| `vibeshield ui` | Open the interactive console |

Aliases: `find` → `search`, `rule` → `rules`, `agent` → `agents`,
`completions` → `completion`, `menu` / `console` → `ui`.

## `scan`

```bash
vibeshield scan .                     # the whole project
vibeshield scan ../api                # another project, using its own config
vibeshield scan --diff main           # only findings on lines added vs main
vibeshield scan --staged              # what the pre-commit hook runs
vibeshield scan --diff -  < patch.diff   # a unified diff from stdin
vibeshield scan . --format json       # machine-readable
vibeshield scan . --format sarif      # GitHub code scanning
vibeshield scan . --mode block-on-critical
vibeshield scan . -v                  # explain what was scanned, on stderr
```

| Flag | Meaning |
|:---|:---|
| `--diff <ref\|->` | Diff mode: scan changes vs a git ref, or `-` to read a diff from stdin |
| `--staged` | Pre-commit mode: scan git staged changes |
| `--format <fmt>` | `pretty` (default) · `json` · `github` · `sarif` |
| `--config <file>` | Config path. Default: `<path>/vibeshield.yml`, then `./vibeshield.yml` |
| `--mode <mode>` | `off` · `warn` · `block-on-critical` · `block-on-high+` |
| `--rules <dir>` | Load extra rule packs from a directory |
| `--online` | (reserved) allow package-intel network lookups |
| `--max-cols <n>` | Output width cap (default 88) |
| `--no-color` | Disable colour (also `NO_COLOR`, and automatic on a non-TTY) |
| `-v`, `--verbose` | Explain what was scanned, on stderr |

## `fix`

```bash
vibeshield fix . --dry-run            # show the −/+ diff, change nothing
vibeshield fix .                      # prompt per file: [y/N/a/s]
vibeshield fix . --yes                # apply everything, no prompts
vibeshield fix . --report scan.json   # reuse a previous --format json scan
```

| Flag | Meaning |
|:---|:---|
| `--dry-run` | Preview the diff, change nothing |
| `--yes` | Apply without prompting; every patch is still logged |
| `--report <file>` | Reuse a `--format json` scan instead of rescanning |
| `--config <file>` | Config path (default `vibeshield.yml` if present) |
| `--rules <dir>` | Load extra rule packs from a directory |
| `--no-color` | Disable colour |
| `-v`, `--verbose` | Explain what was scanned, on stderr |

Every applied patch is appended to `vibeshield-fixes.log` (JSONL: time, rule id,
file, line, before, after, mode). Two categories are never patched
mechanically: a leaked secret needs rotation, and a prompt-injection finding
lives in the very instruction file an agent would be editing.

A non-interactive stdin must pass `--dry-run` or `--yes` — the gate never
guesses at an answer.

## `init`

```bash
vibeshield init --dry-run             # show the plan, write nothing
vibeshield init                       # set the project up and scan it
vibeshield init . --mode block-on-critical
vibeshield init . --no-hook           # CI repo: workflow but no local hook
```

| Flag | Meaning |
|:---|:---|
| `--mode <mode>` | Gate mode written to the config (default `warn`) |
| `--dry-run` | Show what would be written, change nothing |
| `--force` | Overwrite files that already exist |
| `--no-hook` | Skip the git pre-commit hook |
| `--no-workflow` | Skip the GitHub Action workflow |
| `--no-scan` | Skip the first scan |
| `--no-color` | Disable colour |

**Init law** — it writes at most three files (`vibeshield.yml`,
`.github/workflows/vibeshield.yml`, `.git/hooks/pre-commit`), never overwrites
any of them without `--force`, and never touches an existing git hook at all.
With no git repository present it skips the hook and still writes the other
two. The workflow pins the Action to the release the running binary came from,
not to a moving major tag.

## `doctor`

```bash
vibeshield doctor
vibeshield doctor ../api
vibeshield doctor --format json
```

| Flag | Meaning |
|:---|:---|
| `--config <file>` | Config path to check |
| `--format <fmt>` | `pretty` (default) · `json` |
| `--no-color` | Disable colour |
| `-v`, `--verbose` | Print every check, including the ones that passed |

Checks: the binary and its embedded pack, the project path, the config, the git
repository, the pre-commit hook, the PR-gate workflow, and the agent rule
files. Nothing is modified. `--format json` emits `vibeshield.doctor/v1`.

## `search` and `rules`

```bash
vibeshield rules                      # every rule, grouped by category
vibeshield rules VS-SEC-017           # one rule, in full
vibeshield search "prompt injection"  # every token must match
vibeshield search --agents cursor     # agent setup recipes only
vibeshield search --rules --list      # every shipped rule
vibeshield rules --format json        # vibeshield.search/v1
```

| Flag | Meaning |
|:---|:---|
| `--list` | List every catalog entry instead of searching |
| `--rules` | Search rule packs only |
| `--agents` | Search agent setup recipes only |
| `--limit <n>` | Max results (`search` default 20; `rules` default: no limit) |
| `--format <fmt>` | `pretty` (default) · `json` |
| `--no-color` | Disable colour |

The ranker folds separators (`vs-sec-017` ≡ `vs sec 017` ≡ `vssec017`), prefers
word-boundary hits, lets keywords reinforce a title match, and only falls back
to fuzzy subsequence matching for queries long enough to be meaningful — so a
three-letter search returns the AWS rule, not a wall of coincidences.

## `agents`

```bash
vibeshield agents                     # every recipe
vibeshield agents cursor              # one agent
vibeshield agents --body > AGENTS.md  # just the pasteable rule block
vibeshield agents --markdown          # the support matrix as a table
```

| Flag | Meaning |
|:---|:---|
| `--body` | Print only the shared rule block |
| `--markdown` | Print the agent matrix as markdown |

See [agent setup](/docs/quickstart) for the full matrix.

## `completion`

```bash
vibeshield completion bash   >> ~/.bashrc
vibeshield completion zsh    >  "${fpath[1]}/_vibeshield"
vibeshield completion fish   >  ~/.config/fish/completions/vibeshield.fish
vibeshield completion powershell | Out-String | Invoke-Expression
```

The scripts are generated by the binary from the same tables the parser uses,
so they can never offer a flag that no longer exists.

## `version`, `help`, `ui`

```bash
vibeshield version        # also: vibeshield --version, -v, -V
vibeshield help scan
vibeshield ui             # the interactive console, same as a bare vibeshield
```

`vibeshield version` reports the binary version, the embedded rule packs, and
how many rules the engine can actually evaluate:

```console
vibeshield 3.0.0
rule packs: core 2.0.0 (MIT, 122 rules)
engine:     117 active · 5 reserved (structural — pending the package-intel model)
no code leaves this machine: static analysis only
```

The `active` / `reserved` split is deliberate — see
[rules & detection](/docs/rules#coverage).

## Exit codes

| Code | Meaning |
|:---|:---|
| `0` | Clean, or findings present in `warn` / `off` mode |
| `1` | Findings at or above the block threshold in a `block-*` mode |
| `2` | Config or usage error |

`doctor` uses `1` to mean "something needs fixing", because that is what a
health check is for. Asking for help — `-h` on any command — exits `0`.

## JSON output

`--format json` emits the machine report every surface renders from:

```json
{
  "schema_version": 1,
  "tool": "vibeshield",
  "version": "3.0.1",
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

`--format sarif` emits SARIF 2.1.0 for GitHub code-scanning uploads, with a
`partialFingerprints` entry built from the finding's dismiss hash so an alert
survives reformatting. `--format github` emits `::error` / `::warning`
workflow annotations inline in the diff.

## Privacy

Static analysis only. `--online` is reserved and does not make network calls in
this build; when package-intel lands it will send **package names**, never code.
The scanner's default posture is no network. The in-agent
[bug-hunting skill](/docs/skill) is offline by design too.
