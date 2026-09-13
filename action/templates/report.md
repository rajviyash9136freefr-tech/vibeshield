<!-- vibeshield-report:v1 -->
<!--
  VibeCheck report — template for the consolidated PR comment (UIUX.md 4.1).
  entrypoint.sh renders this anatomy from the scanner's `--format json` output
  via an embedded jq program; this file is the human-readable reference and the
  golden shape that action/test/smoke.sh checks against.

  Rules the renderer follows:
  - Severity is always conveyed as TEXT (CRITICAL/HIGH/MEDIUM/LOW/INFO); the
    leading emoji are decorative (UIUX 6 screen-reader note).
  - One collapsible <details> per finding; the header row carries the severity
    chip cluster, clean-file count, and runtime.
  - file:line links, a "Why this matters for AI code" line, and a one-line fix
    per finding (PRD 5.4 format).
  - Dependency delta table from the JSON `dependencies` array
    (package/version/age/maintainers/risk/why).
  - Footer: dismiss command using the finding's dismiss_hash, docs link, badge.
  - Tone: blame patterns, never people. No fear-mongering.
-->

🛡 **VibeCheck Report** — 1 critical · 1 high · 11 clean · ⏱ 1.23s

<sub>Severity is shown as text; the emoji are decorative. Analysis ran on this runner — source code never left it.</sub>

**CRITICAL** 🔴 · `VS-PKG-001` · hallucinated-package — Hallucinated package dependency

<details>
<summary><code>src/parse.ts:12</code> — LLMs invent plausible package names; attackers register them within hours (…)</summary>

- **Why this matters for AI code:** LLMs invent plausible package names; attackers register them within hours ("slopsquatting"). This package is 9 days old, has 1 maintainer, no README, and an install script that curls a remote host.
- **Location:** `src/parse.ts:12` · AI-origin: confirmed
- **Fix:** Remove it — the usage is 12 lines; replace with `node:util` or a vetted package (e.g. `yargs`).
- **Triage:** `/vibeshield accept a1b2c3 --reason "intentional test fixture"`

```
import parseUtils from "fast-parse-utils-v3";
```
</details>

**HIGH** 🟠 · `VS-SEC-017` · hardcoded-secret — API key echoed in generated boilerplate

<details>
<summary><code>src/lib/agent.ts:41</code> — Key pasted into source is likely straight from an AI chat response (…)</summary>

- **Why this matters for AI code:** Keys pasted into source tend to come from training data or a chat response; the value is often live. Rotate it, move it to the environment, and check whether the same key appears elsewhere in the diff.
- **Location:** `src/lib/agent.ts:41` · AI-origin: likely
- **Fix:** Move to env, rotate the key now
- **Triage:** `/vibeshield accept d4e5f6 --reason "intentional test fixture"`

```
const key = "sk-proj-FAKE••••••••••••4a2f";
```
</details>

**Dependency delta** — what this change added to the tree:

| Package | Version | Age | Maintainers | Risk | Why |
| --- | --- | --- | --- | --- | --- |
| `fast-parse-utils-v3` | 2.1.4 | 9d | 1 | critical 🔴 | registered 9 days ago; 1 maintainer; post-install script fetches remote binary |

---

Reply here to triage: `/vibeshield accept <dismiss_hash> --reason "intentional test fixture"` — acceptances are audit-logged.

Docs: https://vibeshield.dev/docs/github-action · Mode: `warn` · Add the README badge: [![VibeShield](https://img.shields.io/badge/VibeShield-2%20findings%20%7C%201%20critical-F87171)](https://vibeshield.dev)

<!-- vibeshield:blame-patterns-not-people -->
