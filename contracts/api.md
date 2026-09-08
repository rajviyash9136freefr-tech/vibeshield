# VibeShield API contract (v1) — api.vibeshield.dev

Fastify + TypeScript, Node 22+, PostgreSQL (findings store), npm/PyPI crawlers for
package-intel. Deploy target: single container. **v1 ships as a runnable service with
an in-memory/pg-lite fallback store so the whole stack works locally without infra.**

Auth: GitHub OAuth (server-side flow). Unauthenticated = package-intel public endpoints only.

## Endpoints

### Package intelligence (public, cache-friendly, the server-side moat)

```
GET /v1/intel/npm/:package          → IntelReport
GET /v1/intel/pypi/:package         → IntelReport
GET /v1/intel/npm/:package?versions=2.1.4,2.1.5   → per-version detail
```

IntelReport:
```json
{
  "name": "fast-parse-utils-v3",
  "ecosystem": "npm",
  "latest_version": "2.1.4",
  "created_at": "2026-08-30T00:00:00Z",
  "age_days": 9,
  "maintainers": 1,
  "maintainer_accounts_created_recently": true,
  "has_install_scripts": true,
  "install_script_commands": ["postinstall"],
  "downloads_monthly": 12,
  "dependents_count": 0,
  "readme_length": 0,
  "repository_url": null,
  "typosquat_distance_to": ["fast-parse-utils", "fast-parser-utils"],
  "namescore": 0.91,
  "risk": "critical",
  "why": "registered 9 days ago; 1 maintainer; post-install script fetches remote binary",
  "signals": [
    { "name": "age", "weight": 0.25, "value": 9, "verdict": "risky" },
    { "name": "maintainers", "weight": 0.2, "value": 1, "verdict": "risky" },
    { "name": "install_scripts", "weight": 0.25, "value": ["postinstall"], "verdict": "risky" },
    { "name": "downloads", "weight": 0.1, "value": 12, "verdict": "unknown" },
    { "name": "namescore", "weight": 0.2, "value": 0.91, "verdict": "risky" }
  ]
}
```

Risk mapping: critical = install_scripts && (age<30 || maintainers<=1) || namescore≥0.85;
high = age<30 || (maintainers<=1 && downloads<100); medium = age<180; low otherwise.
`namescore` = LLM-name-likelihood heuristic (generated-looking name), computed from
lexical features (version-suffix patterns, kebab soups, "utils/core/helper/helper-v2" shape,
training-cutoff presence check). v1: deterministic heuristic, no model file.

### Scans / findings (auth: GitHub OAuth)

```
POST /v1/scans                Ingest a scan result (Action mode). Body = CLI JSON output object.
GET  /v1/scans/:id            One scan + findings
GET  /v1/repos/:owner/:name/scans   History (paginated, ?limit=20&before=<id>)
GET  /v1/repos/:owner/:name/findings?status=open|accepted|dismissed
POST /v1/findings/:hash/dismiss   { reason } → audit-logged
GET  /v1/repos/:owner/:name/audit   Ignore audit log
GET  /v1/health                  → { ok: true, version, pack_versions }
```

Ingest rule: **file contents never accepted or persisted** — findings metadata only
(paths, rule IDs, symbol names, redacted snippets ≤120 chars). Reject bodies containing
keys `file_contents`, `source`, or snippets > 400 chars with 422.

## Storage schema (PG; in-memory fallback implements same interface)

```sql
orgs(id, github_login, created_at)
users(id, org_id, github_login, created_at)
repos(id, org_id, owner, name, created_at)
scans(id, repo_id, github_pr, commit_sha, mode, files_scanned, duration_ms, summary_jsonb, created_at)
findings(id, scan_id, dismiss_hash, rule_id, severity, category, title, message, file, line, ai_origin, confidence, metadata_jsonb, status, created_at)
dismissals(id, finding_hash, actor, reason, created_at)
```

## Non-functional

- All responses: `X-VibeShield-Version` header. Errors: `{ error: { code, message } }`.
- Rate limit intel endpoints: 60/min/IP unauth, 600/min token.
- npm registry adapter must degrade: on registry 5xx/timeout, serve last-known snapshot
  with `stale: true` flag; never fail a scan because intel is down (offline heuristics
  in the scanner cover the air-gapped path).
- Secrets in config via env only. No PII beyond GitHub login.
