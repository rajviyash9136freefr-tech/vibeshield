// Ship-ready copy from PRD.md §7.2 / §8 and UIUX.md §3. The specs are the
// source of truth — keep FAQ text identical here and in the rendered HTML
// (UIUX §9.4: Google cross-checks JSON-LD against visible text).

export const SITE = {
  name: 'VibeShield',
  // GitHub Pages origin — see site/astro.config.mjs (base = /vibeshield/).
  url: 'https://rajviyash9136freefr-tech.github.io',
  title: 'VibeShield — Bug Hunter & Tester for AI Vibe Coding | Cursor, Claude, Copilot',
  description:
    'Test your vibe-coded apps, hunt down hallucinated packages, logic bugs, and secret leaks, and fix them in seconds. Works inside Cursor, Claude Code, GitHub Copilot, and Windsurf.',
  repo: 'https://github.com/rajviyash9136freefr-tech/vibeshield',
};

export const NAV = [
  { label: 'Features', href: '/#features' },
  { label: 'Audit Demo', href: '/#scanner' },
  { label: 'Agent Setup', href: '/#agents' },
  { label: 'Install', href: '/#install' },
  { label: 'Docs', href: '/docs' },
  { label: 'FAQ', href: '/#faq' },
];

export const HERO = {
  h1: 'The Bug Hunter & Tester for AI Vibe Coding',
  sub: 'Test your vibe-coded apps, hunt down hallucinated packages, logic bugs, and secret leaks, and fix them in seconds. Works directly inside Cursor, Claude Code, GitHub Copilot, and Windsurf.',
  trust: ['100% Local & Offline', 'MIT Free Forever', 'Zero Code Sent Outside', 'Instant 1-Click Fixes'],
};

// Hero terminal demo (UIUX §3.2) — the final frame; JS types it in.
// Kept byte-identical in shape to what `vibeshield scan` actually prints, so
// the demo cannot promise an output format the binary does not produce.
export const DEMO_LINES = [
  { text: '$ vibeshield scan .', cls: 'accent' },
  { text: '', cls: 'gap' },
  { text: '  Scanning 14 files (full mode)… done in 1.2s', cls: 'muted' },
  { text: '', cls: 'gap' },
  { text: '  🔴 CRITICAL  VS-PKG-001  hallucinated-package', cls: 'critical' },
  { text: '     package.json:8 — fast-parse-utils-v3@2.1.4 (9d old, 1 maint)', cls: 'fg' },
  { text: '     remote payload install script detected', cls: 'fg' },
  { text: '     → Fix: replace with node:util (12 lines, zero deps).', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  🟠 HIGH      VS-SEC-017  hardcoded-secret', cls: 'high' },
  { text: '     src/agent.ts:41 — OPENAI_API_KEY pasted from chat context', cls: 'fg' },
  { text: '     → Fix: read process.env.OPENAI_API_KEY and rotate it now.', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  ✓ 12 files clean · 2 findings · 0 leaks merged to main', cls: 'ok' },
  { text: '', cls: 'gap' },
  { text: '  Next: vibeshield fix . --dry-run   (preview the one-line fixes)', cls: 'faint' },
];

export const INSTALL_TABS = [
  {
    id: 'script',
    label: '1-Line Install Script',
    code: `# macOS & Linux (Bash):
curl -fsSL https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.sh | bash

# Windows (PowerShell):
irm https://raw.githubusercontent.com/rajviyash9136freefr-tech/vibeshield/main/scripts/install.ps1 | iex

# Then confirm it is wired up:
vibeshield doctor`,
    lang: 'bash',
  },
  {
    id: 'agent',
    label: 'AI Agent Skill',
    code: `# Claude Code:
claude plugin marketplace add rajviyash9136freefr-tech/vibeshield
claude plugin install vibeshield@vibeshield

# Cursor / VS Code / Windsurf:
# Automatic via install script, or copy .cursorrules from the Agent Prompts section above!`,
    lang: 'bash',
  },
  {
    id: 'npx',
    label: 'Zero-Install (npx / npm)',
    code: `# Instant scan without installation:
npx vibeshield scan .

# Or install globally via npm:
npm install -g vibeshield

# Tab completion, once installed:
vibeshield completion bash >> ~/.bashrc`,
    lang: 'bash',
  },
  {
    id: 'action',
    label: 'GitHub Action (CI/CD)',
    code: `# .github/workflows/vibeshield.yml
name: VibeShield PR Gate
on: [pull_request]
jobs:
  audit:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: rajviyash9136freefr-tech/vibeshield/action@v3.0.0`,
    lang: 'yaml',
  },
];

export interface AgentPrompt {
  id: string;
  name: string;
  badge: string;
  filename: string;
  description: string;
  prompt: string;
  cliCommand?: string;
}

export const AGENT_PROMPTS: AgentPrompt[] = [
  {
    id: 'cursor',
    name: 'Cursor',
    badge: 'IDE Rules & Agent',
    filename: '.cursorrules',
    description: 'Paste into your project root as `.cursorrules` to instruct Cursor to hunt and fix AI bugs automatically before saving.',
    prompt: `# VibeShield Bug Hunter Rule for Cursor
You are an adversarial code auditor and bug hunter for vibe-coded applications.
Before finalizing, applying, or committing any code in this project:
1. HALLUCINATION CHECK: Audit every imported package, module, or API. If a library is not well-established, flag it and replace with standard library or verified packages.
2. SECRET SCAN: Ensure NO OpenAI, Anthropic, Stripe, AWS, or database credentials are leaked in client-side code or git diffs.
3. LOGIC & COMPONENT TESTING: Inspect state updates, race conditions, async error handling, and broken edge cases in UI components.
4. INSECURE DEFAULTS: Never scaffold wildcard CORS (*), debug=True, weak session secrets, or JWT algorithms:["none"].
5. ATOMIC FIX: For every bug found, provide an immediate one-line diff replacement.`,
  },
  {
    id: 'claude',
    name: 'Claude Code',
    badge: 'Terminal Skill / Plugin',
    filename: '~/.claude/skills or /plugin',
    description: 'Install the official VibeShield bug-hunting skill or run direct audit prompts in your Claude Code sessions.',
    cliCommand: '/plugin marketplace add rajviyash9136freefr-tech/vibeshield && /plugin install vibeshield@vibeshield',
    prompt: `/vibeshield:vibeshield-audit --quick

Please audit this vibe-coded project:
1. Hunt for hallucinated npm/pip packages and unverified imports.
2. Check for leaked API tokens, hardcoded secrets, and insecure endpoints.
3. Test component logic, broken state flows, and unhandled exceptions.
4. Output a verified VibeCheck report and apply 1-line fixes.`,
  },
  {
    id: 'windsurf',
    name: 'Windsurf (Cascade)',
    badge: 'Cascade Rules',
    filename: '.windsurfrules',
    description: 'Add to `.windsurfrules` in your workspace so Cascade constantly tests output for vibe coding bugs.',
    prompt: `# VibeShield Security & Bug Hunter for Windsurf Cascade
When generating or refactoring code in this repository:
1. Verify all dependencies against hallucinated package names (slopsquatting).
2. Scan all generated code for credentials, API tokens, and private environment variables.
3. Test edge-case logic, error handling, and component state transitions.
4. Reject insecure boilerplate: wildcard CORS, disabled auth, and unsafe evals.
5. Provide safe, production-grade replacements for any identified vulnerability.`,
  },
  {
    id: 'copilot',
    name: 'GitHub Copilot',
    badge: 'Copilot Instructions',
    filename: '.github/copilot-instructions.md',
    description: 'Commit to `.github/copilot-instructions.md` to guide Copilot Chat & Agent PR reviews.',
    prompt: `# VibeShield Bug Hunter Instructions for GitHub Copilot
Act as VibeShield Bug Hunter and adversarial code reviewer for AI-generated diffs:
1. Identify hallucinated packages and unpinned supply chain risks in package manifests.
2. Quarantine any hardcoded secrets, API tokens, or credentials echoed from context.
3. Scrutinize component state, edge cases, and asynchronous error boundaries.
4. Flag insecure defaults (CORS *, debug flags, permissive JWTs) and suggest clean diff fixes.`,
  },
  {
    id: 'terminal',
    name: 'Aider / Terminal CLI',
    badge: 'CLI Command',
    filename: 'Terminal / Bash',
    description: 'Run directly in your terminal or pass to Aider / custom agent scripts.',
    cliCommand: 'npx vibeshield scan --diff main',
    prompt: `npx vibeshield scan --diff main

# Or run full repository bug-hunt audit:
npx vibeshield scan .`,
  },
];

// Exactly 5 Curated FAQs as explicitly requested:
export const FAQ: { q: string; a: string; link?: [string, string] }[] = [
  {
    q: 'What is VibeShield and how does it test vibe-coded apps?',
    a: 'VibeShield is an adversarial bug hunter and security tester built specifically for AI vibe coding. When AI agents (Cursor, Claude Code, GitHub Copilot, Windsurf) generate hundreds of lines of code, they introduce hallucinated packages, leaked keys, and broken component logic. VibeShield inspects code the moment it is written, hunts these specific bugs, and gives you instant 1-click fixes before you commit or merge.',
    link: ['Explore Agent Prompts', '/#agents'],
  },
  {
    q: 'How do I install VibeShield and use it with my AI coding agent?',
    a: 'Four ways, one binary. (1) Run the install script on macOS/Linux or the PowerShell installer on Windows. (2) Copy our agent prompt into .cursorrules, .windsurfrules or the Claude Code plugin. (3) Add the GitHub Action to gate every pull request in seconds. (4) npx vibeshield scan . for a zero-install scan. Then run vibeshield doctor to confirm everything is wired up.',
    link: ['View Installation Options', '/#install'],
  },
  {
    q: 'What specific bugs and vulnerabilities does VibeShield hunt down?',
    a: 'VibeShield targets failure modes unique to LLM-generated code: hallucinated npm/pip packages (slopsquatting attacks where attackers pre-register names invented by LLMs), leaked API keys and passwords echoed in boilerplate, insecure wildcard CORS (*), broken JWT auth, stripped licence headers, and prompt-injection backdoors. 122 rules across 6 categories — all readable with vibeshield rules.',
    link: ['Test with Live Simulator', '/#scanner'],
  },
  {
    q: 'Does my source code or prompt data leave my machine?',
    a: 'No. Never. VibeShield is designed for zero data exfiltration. The CLI, console, pre-commit hook, GitHub Action and agent skill all execute 100% locally using static analysis. There is no telemetry, no account, no dashboard, and no server component — the scanner links in no HTTP client at all. Your code, API keys, and prompts are never sent anywhere.',
    link: ['Read the security model', '/security'],
  },
  {
    q: 'Is VibeShield free, and can I use it for private repositories?',
    a: 'Yes. VibeShield is 100% free and open-source under the MIT license. There are no paid tiers, no seat limits, no tokens, and no account required. You can freely use it on unlimited personal, public, and private commercial repositories forever.',
    link: ['View GitHub Repository', SITE.repo],
  },
];

export const FOOTER = {
  cols: [
    {
      title: 'Navigation',
      links: [
        ['How to Setup', '/#agents'],
        ['How to Install', '/#install'],
        ['How to Scan', '/#how-to-scan'],
        ['Live Demo', '/#scanner'],
        ['FAQ', '/#faq'],
      ] as [string, string][],
    },
    {
      title: 'Documentation',
      links: [
        ['Quickstart', '/docs/quickstart'],
        ['Installation', '/docs/installation'],
        ['CLI reference', '/docs/cli'],
        ['Configuration', '/docs/config'],
        ['Rules & detection', '/docs/rules'],
        ['Troubleshooting', '/docs/troubleshooting'],
      ] as [string, string][],
    },
    {
      title: 'Agents Supported',
      links: [
        ['Cursor (.cursorrules)', '/#agents'],
        ['Claude Code (/plugin)', '/#agents'],
        ['Windsurf (.windsurfrules)', '/#agents'],
        ['GitHub Copilot Gate', '/#agents'],
        ['Terminal / Aider', '/#how-to-scan'],
      ] as [string, string][],
    },
    {
      title: 'Open Source',
      links: [
        ['GitHub Repository', 'https://github.com/rajviyash9136freefr-tech/vibeshield'],
        ['Issue Tracker', 'https://github.com/rajviyash9136freefr-tech/vibeshield/issues'],
        ['Discussions', 'https://github.com/rajviyash9136freefr-tech/vibeshield/discussions'],
        ['Releases & Binaries', 'https://github.com/rajviyash9136freefr-tech/vibeshield/releases'],
        ['MIT License', 'https://github.com/rajviyash9136freefr-tech/vibeshield/blob/main/LICENSE'],
      ] as [string, string][],
    },
  ],
  bottom: '© 2026 VibeShield · Bug Hunter & Tester for AI Vibe Coding · Free & Open Source under MIT License',
};
