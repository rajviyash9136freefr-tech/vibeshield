---
title: Upgrading to v3
description: "v3 is a drop-in upgrade for the CLI: the rule pack is unchanged and no existing command, flag or exit code changed meaning. Here is what is new."
order: 10
---

v3 is a **CLI usability release**. The rule pack stays at `2.0.0`, the scanner
behaves the same way, and nothing you already run breaks. What changed is that
a person who has never seen VibeShield can now find their way around it.

```bash
vibeshield --version     # vibeshield 3.0.0
vibeshield doctor        # confirm the upgrade landed cleanly
```

## Nothing breaks

| Surface | v2 | v3 |
|:---|:---|:---|
| `scan` flags | — | unchanged |
| `fix` flags | — | unchanged |
| `init` flags | — | unchanged |
| Exit codes | `0` / `1` / `2` | unchanged |
| `--format json` schema | `schema_version: 1` | unchanged |
| Rule pack | `core 2.0.0` | `core 2.0.0` |
| Bare `vibeshield` on a pipe | help + exit 2 | unchanged |
| `vibeshield -v` | version | unchanged |

There is no migration step. Update the binary and you are done.

## What is new

### `vibeshield doctor`

A read-only health check for the binary, the config, the git hook, the PR gate
and your agent rule files. Every gap comes with the command that fixes it, and
it exits `1` when something needs fixing so CI can assert the gate is wired up.

```bash
vibeshield doctor
vibeshield doctor --format json
```

### Per-command help

`vibeshield scan --help` now prints `scan`'s own flags and examples instead of
the global manual. Every command has its own help, and `vibeshield help <cmd>`
is the same text.

### Shell completion

```bash
vibeshield completion bash   >> ~/.bashrc
vibeshield completion zsh    >  "${fpath[1]}/_vibeshield"
```

### `vibeshield rules`

The rule packs under a name people reach for. `vibeshield rules` lists them all
by category; `vibeshield rules VS-SEC-017` reads one in full. It is
`search --rules` underneath, so the two can never disagree.

### Typo suggestions

```console
$ vibeshield scna
vibeshield: unknown command "scna"

  Did you mean `vibeshield scan`?
```

### `-V`

`vibeshield -V` is now a version alias alongside `-v` and `--version`. The
`-v`/`--verbose` split inside subcommands is unchanged: `vibeshield -v` is the
version, `vibeshield scan -v` is verbose.

## Bug fixes worth knowing about

- **`search --rules` no longer eats the next flag.** The positional-argument
  rewriter used one global table of value-taking flags, so in `search` — where
  `--rules` is a boolean switch — it consumed the following argument as if it
  were a directory name. `vibeshield search --rules --list` and
  `vibeshield rules VS-SEC-017 --no-color` both failed with a flag parse error
  in v2.
- **`search --list` is no longer truncated.** It was capped at the default
  `--limit 20` even though the docs called it "every entry". An explicit
  `--limit` still applies.
- **`doctor --config <missing-file>`** is a warning, not a silent fall-through
  to the defaults.

## Upgrading the CI pin

Change the ref in your workflow and your pre-commit config:

```yaml
# .github/workflows/vibeshield.yml
- uses: rajviyash9136freefr-tech/vibeshield/action@v3.0.0
```

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v3.0.0
```

`vibeshield init` pins the Action to the release the running binary came from,
so re-running it after the upgrade rewrites the workflow for you — it will skip
the file rather than overwrite it, so pass `--force` if you want it updated in
place.

## Upgrading the agent rule files

The shared rule block is unchanged in v3, so your `AGENTS.md`,
`.cursorrules` and friends do not need touching. If you want the current text:

```bash
vibeshield agents --body > AGENTS.md
```

## Full detail

[CHANGELOG.md](https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/CHANGELOG.md)
has the complete v3.0.0 entry.
