// Ship-ready copy from PRD.md §7.2 / §8 and UIUX.md §3. The specs are the
// source of truth — keep FAQ text identical here and in the rendered HTML
// (UIUX §9.4: Google cross-checks JSON-LD against visible text).

export const SITE = {
  name: 'VibeShield',
  url: 'https://vibeshield.dev',
  title: 'VibeShield — Security Scanner for AI-Generated Code | GitHub Action',
  description:
    'VibeShield audits Cursor, Copilot & Claude Code output for hallucinated packages, leaked secrets and insecure code. Free GitHub Action + pre-commit hook. 3-minute setup.',
  repo: 'https://github.com/rajviyash9136freefr-tech/vibeshield',
};

export const NAV = [
  { label: 'Product', href: '/#product' },
  { label: 'Docs', href: '/docs' },
  { label: 'Blog', href: '/blog' },
  { label: 'FAQ', href: '/#faq' },
];

export const HERO = {
  h1: 'Security for the code your AI writes',
  sub: 'VibeShield audits every Cursor, Copilot, and Claude Code commit — hallucinated packages, leaked secrets, insecure boilerplate — as a GitHub Action, pre-commit hook, or one command. Before it hits main.',
  trust: ['MIT licensed', '3-min setup', 'Runs locally', 'No code leaves your machine'],
};

// Hero terminal demo (UIUX §3.2) — the final frame; JS types it in.
export const DEMO_LINES = [
  { text: '$ npx vibeshield scan', cls: 'accent' },
  { text: '', cls: 'gap' },
  { text: '  Scanning 14 changed files (diff mode)… done in 1.2s', cls: 'muted' },
  { text: '', cls: 'gap' },
  { text: '  🔴 CRITICAL  VS-PKG-001  hallucinated-package', cls: 'critical' },
  { text: '     fast-parse-utils-v3@2.1.4 — registered 9 days ago, 1 maintainer,', cls: 'fg' },
  { text: '     post-install script fetches remote binary.', cls: 'fg' },
  { text: '     → Fix: replace with node:util (12-line change in src/parse.ts)', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  🟠 HIGH      VS-SEC-017  hardcoded-secret', cls: 'high' },
  { text: '     OPENAI_API_KEY echoed in src/lib/agent.ts:41 — likely pasted', cls: 'fg' },
  { text: '     from an AI chat response.', cls: 'fg' },
  { text: '     → Fix: move to env, rotate the key now', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  ✓ 11 files clean · 2 findings · 1 dependency added (risky)', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  Full report → github.com/you/api/pull/482#vibeshield', cls: 'faint' },
];

export const PROBLEM = {
  h2: 'Your AI ships 500 lines a day. Who reads them?',
  paras: [
    'AI agents now write a third to a half of new code in adopting teams. The same developers who would never merge a teammate\x27s unread diff accept a model\x27s output after a five-second glance. The industry calls it vibe coding; the security consequence is simply that the review step went missing.',
    'LLMs fail in specific, enumerable ways. They invent package names that look plausible — and attackers register those names within hours. They echo secrets out of training data. They scaffold debug mode, wildcard CORS, and accept-alg-none JWTs because that code once worked in a tutorial.',
    'Classic scanners were built for a world where a human wrote the code and risk meant known CVEs. A hallucinated package has no CVE. A pasted key has no advisory. A dropped license header leaves no trace. These defects fall between the seats of tools tuned to a different era.',
    'The fix is not more scanning. It is a gate that understands how AI code fails, attributes findings to the hunks the model wrote, and posts them where the decision gets made: the pull request.',
  ],
  // PRD §2.1 failure-mode table — semantic <table> for `is AI generated code safe`
  table: {
    caption: 'AI-code failure modes and how classic scanners treat them',
    cols: ['Failure mode', 'What it looks like', 'Classic scanners'],
    rows: [
      ['Hallucinated packages', 'Model invents `fast-parse-utils-v3`; attacker pre-registers it', 'Not flagged — no CVE yet'],
      ['Secrets in boilerplate', 'API key pasted out of training data or chat context', 'Partial (generic secret tools)'],
      ['Outdated insecure APIs', 'md5 passwords, eval on input, SQL string concat', 'Rule-dependent, noisy'],
      ['License stripping', 'GPL snippet arrives with the header removed', 'Not detected'],
      ['Insecure defaults', 'CORS *, debug=True, JWT algorithms:["none"]', 'Buried in noise'],
      ['Transitive blowout', 'One added library, 200 new packages nobody saw', 'Later, via npm audit — not at PR'],
      ['Prompt-injection changes', 'Malicious README instructs the agent to open a backdoor', 'Nobody covers this'],
    ] as [string, string, string][],
  },
};

export const STEPS = [
  {
    h3: 'Add the GitHub Action',
    body: 'One file, five lines. No account, no agent, no sidecar.',
    code: `name: CI
on: [pull_request]
jobs:
  vibeshield:
    uses: rajviyash9136freefr-tech/vibeshield/action@v1`,
    lang: 'yaml',
  },
  {
    h3: 'Scan AI-generated pull requests',
    body: 'The VibeCheck report lands on the PR: findings, dependency delta, AI-origin tags — in seconds.',
    code: `🛡 VibeCheck Report — 1 critical · 1 high · 11 clean · 1.2s

🔴 VS-PKG-001 hallucinated-package (AI-origin: confirmed)
   fast-parse-utils-v3@2.1.4 — added by copilot-swe-agent
   → Fix: replace with node:util (12-line change)`,
    lang: 'text',
  },
  {
    h3: 'Merge with an audit trail',
    body: 'Accept or dismiss each finding; dismissals are logged with a reason. Add the badge.',
    code: `[![VibeShield](https://vibeshield.dev/badge/passing.svg)](https://vibeshield.dev)`,
    lang: 'markdown',
  },
];

export const FEATURES = [
  {
    title: 'Hallucinated package detection',
    body: 'Scores every new dependency for slopsquatting risk — age, maintainers, install scripts, name-hallucination likelihood.',
    large: true,
    visual: 'deps',
  },
  {
    title: '120 AI-specific rules, 8 languages',
    body: 'A deliberately narrow ruleset — AI failure modes, not general linting.',
    chips: ['VS-PKG-001', 'VS-SEC-017', 'VS-LIC-003', 'VS-DEP-002', 'VS-INJ-001'],
    visual: 'chips',
  },
  {
    title: 'AI-origin attribution',
    body: 'Knows which diff hunks the bot wrote, via bot-author metadata and writing-pattern heuristics.',
    visual: 'hunk',
  },
  {
    title: 'Secrets re-scan',
    body: 'Entropy + known-format checks tuned to training-data echo.',
    visual: 'secret',
  },
  {
    title: 'Diff-speed',
    body: '1.2s median. Diff-only. Offline rules.',
    visual: 'speed',
  },
  {
    title: 'Explain-grade reports',
    body: 'Every finding: why it matters for AI code, the pattern, a one-line fix, and a dismiss flow.',
    visual: 'report',
  },
];

export const INSTALL_TABS = [
  {
    id: 'action',
    label: 'GitHub Action',
    code: `# .github/workflows/vibeshield.yml
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
      - uses: rajviyash9136freefr-tech/vibeshield/action@v1`,
    lang: 'yaml',
  },
  {
    id: 'precommit',
    label: 'pre-commit',
    code: `# .pre-commit-config.yaml
repos:
  - repo: https://github.com/rajviyash9136freefr-tech/vibeshield
    rev: v1.0.0
    hooks:
      - id: vibeshield`,
    lang: 'yaml',
  },
  {
    id: 'cli',
    label: 'CLI',
    code: `# scan the current project
npx vibeshield scan .

# or set up hook + workflow + config in one command
npx vibeshield init`,
    lang: 'bash',
  },
];

export const COMPARISON = {
  h2: 'VibeShield vs. your current stack',
  cols: ['Dependabot', 'Snyk', 'Semgrep', 'Socket', 'Gitleaks', 'VibeShield'],
  rows: [
    { label: 'AI-specific detections', vals: ['❌', '❌', '❌', '⚠️', '⚠️', '✅'] },
    { label: 'Hallucinated-package scoring', vals: ['❌', '❌', '❌', '⚠️', '❌', '✅'] },
    { label: 'Pre-commit hook', vals: ['❌', '❌', '✅', '✅', '✅', '✅'] },
    { label: 'Diff-speed PR scan', vals: ['⚠️', '⚠️', '⚠️', '✅', '✅', '✅'] },
    { label: 'Free tier', vals: ['✅', '❌', '✅', '❌', '✅', '✅'] },
  ] as { label: string; vals: string[] }[],
};

// PRD §7.2 — verbatim. Rendered as <details> AND serialized into FAQPage JSON-LD;
// the two must match character-for-character.
export const FAQ: { q: string; a: string; link?: [string, string] }[] = [
  {
    q: 'What is VibeShield?',
    a: 'VibeShield is a security and dependency auditor built specifically for AI-generated code. It runs as a GitHub Action, a pre-commit hook, and a CLI, catching hallucinated packages, pasted secrets, insecure AI boilerplate, and risky new dependencies the moment your AI agent writes them — before they reach main.',
    link: ['get started', '/#install'],
  },
  {
    q: "Why isn't Dependabot or Snyk enough for AI-generated code?",
    a: 'Because they look for known vulnerabilities, and most AI-code risk isn\x27t "known" yet. When an LLM invents a package name and an attacker registers it the same week, there is no CVE to match — the package *is* brand-new. VibeShield scores every newly-added dependency on supply-chain signals (age, maintainers, install scripts, name-hallucination likelihood) instead of waiting for someone to report abuse. It complements Dependabot; it doesn\x27t replace it.',
    link: ['the failure-mode table', '/#product'],
  },
  {
    q: 'What is slopsquatting (package hallucination)?',
    a: "LLMs confidently cite packages that don't exist. Attackers harvest those invented names, publish malicious versions to npm/PyPI/crates, and wait for a developer's AI agent to `npm install` them — a documented attack class (see the Boston University and Veracode studies on package hallucination). VibeShield's model flags names that look generated and packages that appeared suspiciously recently, and blocks the install path at PR time.",
    link: ['/blog/slopsquatting-explained', '/blog/slopsquatting-explained'],
  },
  {
    q: 'Does VibeShield work with Cursor, Copilot, and Claude Code?',
    a: 'Yes. It doesn\x27t plug into the agents themselves — it gates what they output. Anything written by Cursor, GitHub Copilot, Copilot Workspace, Claude Code, Windsurf, Aider, Devin, or an agent-authored PR (`copilot-swe-agent`, custom bots) is scanned at commit and at PR. Findings are attributed to likely-AI hunks using bot-author metadata and writing-pattern heuristics.',
    link: ['the bug-hunting skill', '/docs/skill'],
  },
  {
    q: 'Will it slow down my CI or flood me with false positives?',
    a: 'Diff-only scanning takes seconds, not minutes — most PRs finish under 10s in the Action and under 1.5s in the pre-commit hook. The ruleset is deliberately narrow (AI failure modes, not general linting): ~120 rules versus thousands in general SAST tools. Default mode is warn; you choose when it can block a merge.',
    link: ['the precision bar', '/security'],
  },
  {
    q: 'Does my source code leave my machine?',
    a: 'No. The CLI and pre-commit hook run fully local analysis. In GitHub Action mode, only findings metadata (file paths, rule IDs, symbol names) is sent to your dashboard — file contents are never persisted, and private-repo analysis results are never used for model training. Ever.',
    link: ['our threat model', '/security'],
  },
  {
    q: 'How much does VibeShield cost?',
    a: 'Nothing. VibeShield is fully open source under the MIT license — the scanner, the GitHub Action, the pre-commit hook, and every rules pack. No tiers, no seats, no credit card, no account. Install takes one YAML file, and it works offline forever.',
    link: ['the repo', SITE.repo],
  },
  {
    q: 'How do I get started in 3 minutes?',
    a: 'Add `uses: rajviyash9136freefr-tech/vibeshield/action@v1` to your workflow (or run `npx vibeshield init` for the hook + workflow + config in one command). VibeShield scans the next PR, comments a VibeCheck report, and you can enable the badge on your README. No account, no signup — it is MIT-licensed and free for every repo.',
    link: ['install', '/#install'],
  },
  {
    q: 'Can VibeShield fix the code automatically?',
    a: 'v1 explains and suggests — every finding includes a one-line fix — but never rewrites your code. Auto-fix ("VibePatch") is on the roadmap behind an explicit opt-in with diff preview, because trusting an AI to fix AI code without a human gate is exactly the problem we\x27re here to reduce, not reinvent.',
    link: ['the roadmap', '/docs'],
  },
  {
    q: 'Who is responsible for the security of AI-generated code?',
    a: 'Legally and practically, still you — the shipping team. Standards bodies and enterprise procurement are converging on "AI-authored changes need the same gate as human-authored changes"; VibeShield gives you that gate plus the audit trail (who shipped what, when, with what risk) that questions like this require.',
    link: ['the audit trail', '/#how'],
  },
];

export const FOOTER = {
  cols: [
    {
      title: 'Product',
      links: [
        ['Docs', '/docs'],
        ['Changelog', '/changelog'],
        ['Roadmap', '/docs#roadmap'],
        ['Status', 'https://status.vibeshield.dev'],
      ] as [string, string][],
    },
    {
      title: 'Resources',
      links: [
        ['Blog', '/blog'],
        ['Threat model', '/security'],
        ['Slopsquatting guide', '/blog/slopsquatting-explained'],
        ['Bug-hunting skill', '/docs/skill'],
      ] as [string, string][],
    },
    {
      title: 'Compare',
      links: [
        ['vs Snyk', '/compare/snyk'],
        ['vs Socket', '/compare/socket'],
        ['vs Dependabot', '/compare/dependabot'],
      ] as [string, string][],
    },
    {
      title: 'Company',
      links: [
        ['GitHub', 'https://github.com/rajviyash9136freefr-tech/vibeshield'],
        ['X', 'https://x.com/vibeshield'],
        ['Contact', 'mailto:hi@vibeshield.dev'],
        ['Privacy', '/privacy'],
      ] as [string, string][],
    },
  ],
  bottom: '© 2026 VibeShield · MIT open source',
  langs: ['en', 'es', 'de', 'fr'],
};
