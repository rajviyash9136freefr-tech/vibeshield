---
title: GitHub Action
description: "vibeshield/action@v3.0.1 — inputs, outputs, gate modes, PR comments, and the badge. The security gate for AI-generated pull requests."
order: 2
---

`rajviyash9136freefr-tech/vibeshield/action@v3.0.1` is a composite GitHub Action that
scans your PR diff and posts a consolidated VibeCheck report comment. It runs the
same Go scanner binary used by the [pre-commit hook](/docs/pre-commit) and the
[CLI](/docs/cli) — static analysis on the runner, in seconds.

## Minimal workflow

```yaml
name: VibeShield
on: [pull_request]
jobs:
  vibeshield:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # needed for diff mode against the base branch
      - uses: rajviyash9136freefr-tech/vibeshield/action@v3.0.1
```

## Inputs

| Input | Default | Meaning |
|---|---|---|
| `mode` | `warn` | Gate policy: `off` \| `warn` \| `block-on-critical` \| `block-on-high+`. Block modes fail the step (exit 1) when findings at or above the threshold exist. |
| `github_token` | `${{ github.token }}` | Used to create/update the VibeCheck PR comment. |
| `config` | `vibeshield.yml` | Path to the config file. Missing default = scanner defaults; missing explicit path = config error (exit 2). |
| `online` | `false` | Allow network calls for package-intel lookups (`--online`). Degrades gracefully offline. |
| `scanner_bin` | — | Path to a prebuilt binary; skips the release download. |
| `version` | `v3.0.1` | Pinned scanner release tag to download. |

## Outputs

`critical` `high` `medium` `low` `info` `total` — finding counts per severity;
`blocked` — `"true"` when the threshold was met; `report_path` — the rendered
markdown comment on the runner; `report_json` — the raw scan JSON
([contract](/docs/cli#json-output)).

Example — fail the build on criticals and echo the count:

```yaml
- uses: vibeshield/action@v3.0.1
  id: vs
  with:
    mode: block-on-critical
- run: echo "criticals=${{ steps.vs.outputs.critical }}"
```

## The VibeCheck report comment

One consolidated comment per PR, updated in place on every push:

- **Header** — severity chip cluster (`1 critical · 1 high · 11 clean`) + runtime.
- **Findings** — collapsible blocks in the [findings format](/docs/quickstart#what-you-get-per-finding),
  with `file:line` links and `AI-origin: likely|confirmed` tags where the
  hunk attribution says so (bot authors, co-author trailers, generated markers).
- **Dependency delta** — package / version / age / maintainers / risk / why-flagged
  for everything the PR added to the tree.
- **Footer** — dismiss commands, e.g.
  `/vibeshield accept VS-SEC-017 --reason "test fixture"`. Dismissals are logged.

## AI attribution

The Action does not plug into Cursor, Copilot, or Claude Code — it gates what
they output. Hunks are attributed `confirmed` when the commit author is a known
bot (`copilot-swe-agent`, `*[bot]`), a co-author trailer names an agent, or a
generated-marker comment is present; `likely` from writing-pattern heuristics;
`unknown` otherwise.

## Badge

Add the shield to your README once your first scan is green:

```markdown
[![VibeShield](https://rajviyash9136freefr-tech.github.io/vibeshield/badge/passing.svg)](https://github.com/rajviyash9136freefr-tech/vibeshield)
```

Swap `passing` for `findings`, `critical` or `failing` to show the current
state — the four files ship in the repository under `site/public/badge/`. It is
honest data, and a red badge is a feature.

## Privacy

The Action runs the scanner on your runner. The only thing that leaves it is
the VibeCheck comment, posted to your own pull request with your own
`GITHUB_TOKEN`. There is no VibeShield server, no dashboard, and no findings
API — file contents never leave the runner because there is nowhere for them to
go. Details on the [security page](/security).
