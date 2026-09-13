---
title: "Slopsquatting explained: how hallucinated npm packages become malware"
description: "LLMs invent package names; attackers register them. What package hallucination is, the research behind it, and how to block the install path at PR time."
pubDate: 2026-09-08
author: "VibeShield"
tags: ["slopsquatting", "supply-chain", "packages"]
---

## The attack in one paragraph

A language model is asked to parse a CLI flag. Confidently, it writes
`npm install fast-parse-utils-v3` — a package that does not exist. A developer
accepts the diff. Sometimes the install fails and someone removes the line.
Sometimes it doesn't: because between the model inventing the name and the
developer typing `npm install`, someone else registered it. That someone
harvests thousands of hallucinated names from public model output, publishes a
malicious package under each, and waits. The class has a name now —
**slopsquatting**, or package hallucination — and a body of research behind it:
the Boston University study that estimated over 5,000 packages were vulnerable
to hallucinated-name squatting at the time of measurement, and Veracode's
work showing roughly one in five AI-generated code snippets references a
nonexistent package.

[Get started: add VibeShield to your PRs →](/#install) ·
[Compare VibeShield vs Socket →](/compare/socket)

## Why classic tools miss it

A dependency auditor asks: *does this package have known vulnerabilities?*
A slopsquatted package has none — it was registered nine days ago and its
first release is also its last. There is no CVE because the abuse *is* the
existence. The signal lives somewhere else entirely:

| Signal | What "normal" looks like | What a squatted package looks like |
|---|---|---|
| Registry age | years | days — often days after the model's answer |
| Maintainers | 2–50+ | 1, newly-created account |
| README | documents an API | absent, or a copy-paste of the real package it sounds like |
| Install hooks | none | `postinstall: curl … \| sh` |
| Name shape | memorizable, human | morphologically "generated": adjective-verb-utils-v3 |

The last row matters more than people expect. Hallucinated names have a
statistical shape — heavy on plausible-but-unusual token joins, light on
memorized brand names — because they come from a distribution, not from
memory. That shape is scorable.

## What gates actually work

**At generation time:** nothing yet ships reliably; agents install what they
propose. **At commit time:** a pre-commit hook that scores every dependency
*added* to a manifest, offline, in about a second and a half. **At PR time:**
a CI job that diffs the dependency tree, looks up registry signals for new
nodes only, and comments the risky additions with the evidence — age,
maintainers, install scripts, name score — before anything merges.

The rule of thumb we ship in the core pack: *any dependency added by an
AI-authored commit that is younger than 30 days, or has an install script, or
scores as name-generated, needs a human to say its name out loud before it
installs.* That's `VS-PKG-001` and friends — [the full rule list is in the
docs](/docs/rules).

## Practical fixes, in order of value

1. **Gate the diff, not the repo.** New dependencies are the attack surface;
   scanning the whole tree monthly catches yesterday's news.
2. **Pin and vendor.** A lockfile diff is a reviewable artifact. "The agent
   added 1 library and 200 packages" should be visible in the PR, which means
   using `--no-save`-style discipline or a lockfile check in CI.
3. **Never let an install-script run unreviewed in CI.** `npm ci --ignore-scripts`
   by default; opt in per package with a reason.
4. **Treat typosquats and squats as the same bug class.** Edit distance from a
   popular package is one more feature in the score.

VibeShield's package-risk model does 1–2 and scores 3–4 at PR time —
[free for public repos, one YAML file to install](/#install).
