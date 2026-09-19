---
title: Troubleshooting
description: "Fixes for the problems people actually hit: nothing found, exit code 1 or 2, a blocked commit, a red Action, missing completions."
order: 3
---

Start here:

```bash
vibeshield doctor
```

It reports the binary, the config, the git hook, the PR gate and your agent
rule files, and prints the exact command that fixes each gap. It exits `1` when
something needs fixing, so it works in CI too.

If that comes back clean and the problem is still there, find your symptom
below.

## "It found nothing — is it even working?"

The most common cause is that there is genuinely nothing to find: the core pack
is narrow on purpose. Check what it is looking for:

```bash
vibeshield rules                 # the whole pack, by category
vibeshield scan . --format json  # confirm files_scanned is not 0
```

`vibeshield scan . -v` prints the resolved config, mode, languages and rule
count to stderr, which tells you whether a config is quietly narrowing the
scan.

Two real cases where zero findings is expected:

- **Five rules are reserved.** They load and validate so packs stay portable,
  but the engine cannot evaluate them yet, so they never fire. `vibeshield
  version` reports `117 active · 5 reserved`, and `vibeshield rules` marks each
  one `RESERVED`. A fixture built around a hallucinated package will come back
  clean for exactly this reason.
- **`languages:` is filtering.** If your `vibeshield.yml` lists languages, only
  those files are walked.

## Exit code 1

The mode is a block mode and findings at or above the threshold were present.
That is the gate working, not a crash.

```bash
vibeshield scan . --mode warn      # report without failing
```

To let a specific finding through, add an auditable ignore:

```yaml
# vibeshield.yml
ignore:
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"
```

`reason` is required — it is echoed in the PR report and the audit log, so a
dismissal leaves a trail instead of a mystery.

## Exit code 2

Usage or config error. The message on stderr says which. The usual suspects:

| Message | Cause |
|:---|:---|
| `bad mode "..."` | `mode` must be `off`, `warn`, `block-on-critical` or `block-on-high+` |
| `notifications.slack must be an environment reference` | A literal webhook URL in the config. Use `${SLACK_WEBHOOK}` |
| `invalid YAML` | A malformed `vibeshield.yml` — the message names the file |
| `unknown --format "..."` | `pretty`, `json`, `github` or `sarif` |
| `is not a readable directory` | The path does not exist, or is a file |

## `vibeshield scan --diff main` says diff mode is unavailable

Diff mode shells out to git, so it needs three things: a git repository, the
ref you named, and enough history to compare against.

```bash
git rev-parse --verify main          # does the ref exist?
git fetch origin main                # a fresh clone may not have it yet
vibeshield scan --diff HEAD~1        # a ref that always exists locally
```

In CI, `actions/checkout@v4` needs `fetch-depth: 0` for diff mode to see the
base branch. A shallow clone is the single most common cause of this one.

## The pre-commit hook is not running

```bash
ls -l .git/hooks/pre-commit          # does it exist, and is it executable?
```

If it is missing, `vibeshield init` will write it. If a hook is already there,
`init` deliberately leaves it alone — add the scan to it yourself, or:

```bash
vibeshield init --force
```

If you use the `pre-commit` framework instead, check `.pre-commit-config.yaml`
has the hook and that you ran `pre-commit install`.

Remember `git commit --no-verify` always works. VibeShield is a gate you own,
not a hostage-taker.

## The pre-commit hook runs but never finds anything

The hook runs `vibeshield scan --staged`, which only looks at staged changes.
If you have not staged anything, there is nothing to scan:

```bash
git add -A && vibeshield scan --staged
```

## The GitHub Action is red but the local scan is green

The Action scans the **diff**, so it only reports findings on lines the PR
added. Differences to check first:

- **`mode`.** The Action's default is `warn`; your config may set
  `block-on-critical`, or the other way round. The workflow input overrides the
  config file.
- **`config` path.** The input is relative to the workspace root. A path that
  does not exist is a config error (exit 2) — visible in the step log.
- **`fetch-depth: 0`.** Without it the base branch is missing and the Action
  cannot build a diff.

## The Action cannot comment on the PR

The workflow needs `pull-requests: write`:

```yaml
permissions:
  contents: read
  pull-requests: write
```

On a pull request from a fork, `GITHUB_TOKEN` is read-only, so the comment
cannot be posted. The scan and the exit code still work — only the comment is
skipped. Use `pull_request_target` if you need comments from forks, and read
GitHub's guidance on that trigger before you do, because it runs with elevated
permissions.

## Tab completion is not working

```bash
vibeshield completion bash >> ~/.bashrc && source ~/.bashrc
```

For zsh the file must be on `$fpath` and named `_vibeshield` — the directory
matters, not just the filename. `vibeshield completion zsh` prints the exact
install line for your shell at the top of the script.

If you installed via `npx`, completion works but is slower, because each Tab
press goes through the Node launcher. Install the binary directly instead.

## `fix` refuses to run

```console
vibeshield: stdin is not a terminal — refusing to guess at the gate.
```

`fix` prompts per file, and a piped stdin would auto-answer "no" to
everything — which looks like a hang. Choose explicitly:

```bash
vibeshield fix . --dry-run    # preview
vibeshield fix . --yes        # apply, no prompts
```

## Colours look wrong, or appear in a log file

Colour is disabled automatically when stdout is not a terminal, and when
`NO_COLOR` is set. To force it off:

```bash
vibeshield scan . --no-color
NO_COLOR=1 vibeshield scan .
```

## The console does not open

The interactive console needs a real terminal on stdin. Over a pipe — in CI, in
a script, in some IDE terminals — `vibeshield` prints help and exits `2`
instead, so automation is never blocked. Use `vibeshield search` for the same
index on stdout, or `vibeshield ui` to ask for the console explicitly.

## Nothing here matches

- [Open an issue](https://github.com/rajviyash9136freefr-tech/vibeshield/issues)
  — include the output of `vibeshield doctor` and `vibeshield version`.
- [Ask in Discussions](https://github.com/rajviyash9136freefr-tech/vibeshield/discussions)
  if you are not sure it is a bug.
