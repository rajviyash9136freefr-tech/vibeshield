# The bug-hunt playbook

Per-category detection knowledge for `vibeshield-audit`. Load only the sections
your work unit needs. Each section: **grep for → read to confirm → common
false positives**.

## FINDING-SCHEMA (contracts/finding/schema.json — exact law)

```json
{
  "schema_version": 1,
  "rule_id": "VS-SEC-017",
  "severity": "critical | high | medium | low | info",
  "category": "hallucinated-package | hardcoded-secret | insecure-api | license-missing | insecure-default | dependency-risk | prompt-injection",
  "title": "≤ 80 chars, sentence case, names the pattern",
  "message": "why this matters for AI code — 1–3 sentences",
  "file": "repo/relative/posix/path.ts",
  "line": 41, "end_line": 41, "column": 5,
  "snippet": "matched line, secrets redacted: 4-char prefix + 4 middle-masked with • + 4-char suffix",
  "fix": "one imperative line",
  "ai_origin": "confirmed | likely | unknown",
  "confidence": 0.9,
  "dismissable": true,
  "dismiss_hash": "sha256(rule_id + \\n + file + \\n + line(int32le) + \\n + normalized_snippet)[:12 hex]",
  "metadata": {}
}
```

`normalized_snippet` = trimmed, lowercased, whitespace-collapsed.
`ai_origin`: `confirmed` = bot author / `Co-authored-by: <bot>` / generated-marker
in the hunk or file header; `likely` = the surrounding code shows LLM tells
(comments explaining trivial code, `// Note:` style, docstring voice, copy-paste
tutorial idioms); otherwise `unknown`. No extra top-level keys.

## VS-PKG — hallucinated packages (manifests, lockfiles, Dockerfiles, notebooks)

**Grep:** `package.json`/`requirements*.txt`/`pyproject.toml`/`go.mod`/`Gemfile`/`Cargo.toml`/`pom.xml` changed vs baseline (`git diff` if in scope). Names matching generated-shape: `\b[a-z]+-(parse|utils?|fast|async|json)-[a-z0-9]+(-v?\d+)?\b`, or a dep name that is also a sentence ("easy-install-…"), or distance ≤ 2 from a famous package (typosquat).
**Also:** `"scripts": { "postinstall"|"preinstall"|"install": … curl|wget|sh|node -e }`; `pip install` with no version pin in docs/CI; `--unsafe-perm`; git+http deps from new accounts.
**Confirm:** for a suspect name — does the import at its call site look like the package's real API (hallucinations rarely match)? Age/maintainers can't be looked up offline: note "9 days"-style claims only from lockfile/resolution evidence; else mark `--online` needed in `metadata`.
**False positives:** your own packages; workspace-internal names; obvious fixtures (skip, or LOW in tests).

## VS-SEC — secrets & insecure API (all code)

**Grep (secrets):** `(sk|pk)-(live|test|proj)-[A-Za-z0-9-]{10,}`, `gh[pousr]_[A-Za-z0-9]{30,}`, `(AKIA|ASIA)[0-9A-Z]{16}`, `xox[baprs]-`, `glpat-`, `(api_?key|secret|token|password)\s*[:=]\s*["'][^"']{12,}["']`, `-----BEGIN (RSA |OPENSSH )?PRIVATE KEY-----`, JWTs `eyJ[A-Za-z0-9_-]{10,}\.`. Entropy-check bare strings ≥ 24 chars assigned to *key/token/secret* names.
**Grep (insecure API):** `md5(` / `sha1(` near `password`; `eval(` / `exec(` on request/env input; string-concat/f-string/`${…}` into `SELECT|INSERT|UPDATE|DELETE`; `MODE_ECB|mode=ecb`; `verify=False|rejectUnauthorized.*false|InsecureSkipVerify`; `Math.random(` for tokens; `DES|RC4|RC2`; `subprocess…shell=True` with interpolated input; `pickle.loads|yaml.load(` (non-safe) on untrusted bytes; `torch.load(` without `weights_only`.
**Confirm:** the value is a literal, not `process.env`/config lookup; the concatenated variable is externally reachable (trace one level); keys: is it `FAKE`/`example`/`changeme`? Then LOW, `metadata.looks_fake`.
**False positives:** test fixtures with obviously-fake keys (the repo's own fixtures use `sk-proj-FAKE…`); lockfile hashes; placeholder docs; UUIDs in state files.

## VS-LIC — license stripping (new source files, vendored dirs)

**Grep:** files whose first lines reference another project's copyright in comments but carry no license; repo files with SPDX headers while a sibling added file lacks one; `LICENSE`/`COPYING` text pasted into source comments (scraped-header smell). Copyleft-in-permissive-project: a `GPL` mention + the project is MIT.
**Confirm:** read the file header vs `git log --diff-filter=A` for who added it (AI-attributed commits raise `ai_origin`).
**False positives:** header-only-notice styles the project already uses.

## VS-DEP — dependency risk (manifests, CI, Dockerfiles)

**Grep:** added ranges `^"?\w+": ?"[\^~]?(\d+\.)?x|\*|latest"`; `go get`/`pip install` unpinned in CI & Dockerfiles; lockfile line-count delta ≫ manifest delta (transitive blowout — count `resolved|integrity` occurrences added).
**Confirm:** count added lockfile entries ≥ 50 → report the number in `message`.
**False positives:** monorepo workspace churn; lockfile format migrations (whole-file rewrite ≠ blowout; compare entry counts).

## VS-INJ — prompt-injection traps (READMEs, issues, docs, .github/, .claude/, AGENTS.md/CLAUDE.md diffs, comments in code)

**Grep:** `ignore (all |previous |prior |any )?instructions`, `disregard .*rules`, `you (must|should) now`, `do not (tell|mention|reveal) (the )?user`, `curl .*\|\s*(ba)?sh`, `cat .*(\.env|credentials|\.aws)`, base64 blobs adjacent to `echo|eval`, "system"-voice imperatives inside quoted content, instructions addressed to bots in issue/PR templates ("when the AI agent reads this…"), and any file that adds a hook/MCP server or grants itself tools (`allowed-tools`) it didn't need.
**Severity:** an executable payload (exfil, curl|sh) = CRITICAL; manipulation text without payload = MEDIUM; suspicious phrasing = LOW/INFO.
**False positives:** your own docs legitimately *describing* injection (defensive content — quote-attribution is the tell: it explains attacks, not commands them); tutorials.

## Attribution cheat-sheet

`confirmed`: committer `*[bot]`, `copilot-swe-agent`, `cursor-bot`, `claude`/`Co-authored-by: Claude`, `Generated-by:` trailers, `AI-generated` markers in-file.
`likely` (need ≥ 2): comment voice explaining trivial lines; `Note:`/`It's important to note`; perfect tutorial structure with unused flexibility; identifiers from training-era memes (`foo_service`-adjacent novelty); docstrings that rewrite the signature; unused imports matching a package you just flagged.
Otherwise `unknown`. Never attribute by blame-of-person; never name a vendor in user-facing copy beyond tool-compat naming.
