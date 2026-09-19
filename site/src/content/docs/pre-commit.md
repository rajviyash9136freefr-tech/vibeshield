---
title: Pre-commit hook
description: "Scan staged files for AI-code failure modes before every commit. Under 1.5s typical, fully offline, and never a hostage."
order: 3
---

The pre-commit hook runs the same rules engine as the
[GitHub Action](/docs/github-action) locally on your staged files, before
anything reaches a remote. Typical runtime: under 1.5 seconds.

## Install

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v3.0.0
    hooks:
      - id: vibeshield
```

Then:

```bash
pre-commit install
```

No pre-commit framework? The hook is also a plain git hook — `vibeshield init`
writes `.git/hooks/pre-commit` for you alongside the config and workflow files.

Not sure whether it is installed? `vibeshield doctor` reports it, along with the
config and the PR gate.

## What it scans

Only **staged changes** (`git diff --cached`), so the cost tracks your edit
size, not your repo size. Rules run offline from the embedded MIT core pack —
the scanner contains no HTTP client at all, so a missing network is never a
failed commit. `--online` is reserved for a future package-intel lookup and is
a no-op in this build.

## Exit behavior

| Situation | Result |
|---|---|
| Clean, or findings under the block threshold | commit proceeds |
| Findings at/above threshold **and** `mode` is a block mode | commit blocked, findings printed |
| Config error | commit proceeds with a loud warning (a broken config must not hold your repo hostage) |

Default mode is `warn`: findings print, commit proceeds. `--no-verify` always
works. VibeShield is a gate you own, not a hostage-taker.

## Output

Findings print in the VibeCheck format — severity, rule ID, why it matters for
AI code, one-line fix — capped at 88 columns, colours from the severity palette,
plain text when piped.

## Keeping it fast

A hook that is slow gets deleted. Three things keep it honest:

- Only staged files are read, so the cost tracks the diff, not the repo.
- The rules run from the pack embedded in the binary — no file I/O to load them.
- `--staged` uses `git diff --cached`, not a full walk.

If your repository is large enough that even that is too slow, scope the hook
with `files:` in `.pre-commit-config.yaml` so it only runs on the paths you
actually change.
