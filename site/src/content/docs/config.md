---
title: Configuration
description: "vibeshield.yml — mode, language filters, auditable ignores with required reasons, thresholds, and notifications."
order: 5
---

Drop a `vibeshield.yml` at your repo root (or point `--config` at one). Every
key is optional — the defaults are `mode: warn`, all languages, no ignores.

```yaml
mode: warn                 # off | warn | block-on-critical | block-on-high+
languages: [typescript, python]   # filter; omit or empty = all
ignore:
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"   # REQUIRED, echoed in reports
thresholds:
  new_dependency_max_age_days: 30
notifications:
  slack: ${SLACK_WEBHOOK}   # env interpolation only — never a literal secret
```

## mode

| Value | PR check | Pre-commit | Findings posted? |
|---|---|---|---|
| `off` | passes | passes | no |
| `warn` (default) | passes | proceeds | yes |
| `block-on-critical` | fails on ≥ 1 critical | blocks those commits | yes |
| `block-on-high+` | fails on high or critical | blocks | yes |

The mode is a policy you can ratchet: `warn` for a month, then flip to
`block-on-critical` once the noise is gone.

## ignore

Every ignore needs a `reason` — it is stored in the audit log and shown in the
PR report, so dismissals leave a trail instead of a mystery. `paths` are
doublestar globs.

The Action prints a triage line under each dismissable finding, e.g.
`/vibeshield accept <hash> --reason "test fixture"`. Replying with it on the PR
writes the same record the config would, so triage and config never diverge.
(That is a pull-request comment command, not a CLI verb.)

## thresholds

`new_dependency_max_age_days` (default 30) is the supply-chain window: a
newly-added package younger than this gets full scrutiny from the
package-risk model. Lower it to 0 to skip age checks (not recommended).

## notifications

`notifications.slack` reads from the environment — a literal webhook URL in the
file is a config error (exit 2), because a config file in git is a bad home for
a credential. That's the whole philosophy, in one line.
