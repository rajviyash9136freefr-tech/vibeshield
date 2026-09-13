// /compare/* — P1 SEO cluster (PRD §7.1). Honest ✅/⚠️/❌ only; the table
// speaks, the copy doesn't trash-talk. Data mirrors PRD §2.2–2.3 + §2.1 table.
export interface Competitor {
  slug: string;
  name: string;
  tagline: string; // what it does well — genuinely
  summary: string; // why AI-code risk still falls through it
  rows: { label: string; us: string; them: string; note?: string }[];
}

export const COMPETITORS: Competitor[] = [
  {
    slug: 'snyk',
    name: 'Snyk',
    tagline: 'Mature CVE-first dependency and SAST scanning with a huge advisory database.',
    summary:
      'Snyk answers "does this package have known vulnerabilities?" — the right question for human-era risk. Most AI-code failure modes have no CVE yet: a hallucinated package registered nine days ago, a secret echoed from training data, a dropped license header. Run both: Snyk for known vulns, VibeShield at the moment the AI writes the diff.',
    rows: [
      { label: 'Known-CVE database', us: '❌ by design', them: '✅ best in class', note: 'We complement it, not compete with it.' },
      { label: 'Hallucinated-package scoring', us: '✅ dedicated model', them: '❌' },
      { label: 'AI-origin hunk attribution', us: '✅', them: '❌' },
      { label: 'Prompt-injection detection', us: '✅', them: '❌' },
      { label: 'Pre-commit hook', us: '✅', them: '❌' },
      { label: 'Median PR scan time', us: '1.2 s (diff-only)', them: 'minutes (full audit)' },
      { label: 'Free & open source', us: '✅ MIT, every repo', them: '⚠️ limited tier' },
      { label: 'Pricing', us: 'free · MIT open source', them: '$$(per developer)' },
    ],
  },
  {
    slug: 'socket',
    name: 'Socket',
    tagline: 'Deep supply-chain behavior analysis of packages — the closest neighbor to our package model.',
    summary:
      'Socket actually reads what packages do (install hooks, network calls, binary fetches) — that\'s real signal, and we\'d point customers at it too. What it doesn\'t do: know that the package was *invented by an LLM this morning* (name-generation likelihood, registry-age-vs-training-cutoff), scan the code itself, or attribute risky hunks to the bot that wrote them. Its behavior feed and our hallucination model look at the same tree from two directions.',
    rows: [
      { label: 'Package behavior analysis', us: '⚠️ install-script rules', them: '✅ dedicated' },
      { label: 'Hallucinated-name scoring', us: '✅ namescore model', them: '⚠️ heuristic' },
      { label: 'Code scanning (secrets, insecure APIs)', us: '✅ ~120 rules, 8 languages', them: '❌' },
      { label: 'AI-origin attribution', us: '✅', them: '❌' },
      { label: 'License stripping', us: '✅ VS-LIC rules', them: '❌' },
      { label: 'GitHub Action + pre-commit + CLI', us: '✅', them: '✅' },
      { label: 'Free & open source', us: '✅ MIT, every repo', them: '⚠️ team-priced' },
    ],
  },
  {
    slug: 'dependabot',
    name: 'Dependabot',
    tagline: 'GitHub-native update and CVE alerts. Free, and it keeps working with VibeShield on.',
    summary:
      'Dependabot waits for an advisory, then opens an update PR. That loop assumes somebody eventually finds and reports the vulnerability. A slopsquatted package is abuse reported zero weeks ago, and the AI-specific problems — pasted secrets, debug-mode scaffolding, injection payloads — aren\'t dependencies at all. Is Dependabot enough? For its lane: yes, keep it. It just covers one of the seven failure modes.',
    rows: [
      { label: 'CVE update PRs', us: '❌ out of scope', them: '✅ native, free' },
      { label: 'Hallucinated packages (no CVE yet)', us: '✅', them: '❌ nothing to match' },
      { label: 'Secrets / insecure boilerplate in code', us: '✅', them: '❌' },
      { label: 'Transitive tree delta at PR time', us: '✅ per-PR', them: '⚠️ alerts after merge' },
      { label: 'AI-origin attribution', us: '✅', them: '❌' },
      { label: 'Works alongside the other', us: '✅', them: '✅' },
    ],
  },
];
