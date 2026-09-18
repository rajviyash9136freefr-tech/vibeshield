#!/usr/bin/env node
// Keeps every agent rules file byte-identical to the Go constant the console
// and `vibeshield agents --body` print. The Go source is the single source of
// truth: edit AgentRulesBody there, then run this.
//
//   node scripts/sync-agent-rules.mjs           # write the files
//   node scripts/sync-agent-rules.mjs --check   # CI: fail on drift
//
// Windows-first (Git Bash), plain Node, no dependencies.
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = fileURLToPath(new URL('..', import.meta.url));
const source = join(repoRoot, 'scanner', 'internal', 'cli', 'catalog.go');

// --- extract the shared rule body -------------------------------------------

function extractBody() {
  const src = readFileSync(source, 'utf8');
  const marker = 'const AgentRulesBody = `';
  const start = src.indexOf(marker);
  if (start === -1) throw new Error(`AgentRulesBody not found in ${source}`);
  const bodyStart = start + marker.length;
  const end = src.indexOf('`', bodyStart);
  if (end === -1) throw new Error('AgentRulesBody raw string is unterminated');
  return src.slice(bodyStart, end);
}

const BODY = extractBody();

// The body lives inside a Go raw string literal, so a stray backtick would
// silently truncate it. Fail loudly here rather than at the next Go build.
if (BODY.includes('`')) {
  console.error('error: AgentRulesBody contains a backtick and would break the Go build');
  process.exit(1);
}

// --- targets ----------------------------------------------------------------
// Each target wraps the shared body in the frontmatter its client needs.
// Cursor's current rules format is MDC and takes YAML frontmatter; the legacy
// .cursorrules / .windsurfrules files are plain markdown.

const mdcFrontmatter = [
  '---',
  'description: VibeShield bug hunter — blocks hallucinated packages, leaked',
  '  secrets, and insecure AI defaults before they are committed.',
  'globs: ["**/*"]',
  'alwaysApply: true',
  '---',
  '',
  '',
].join('\n');

const targets = [
  { file: 'AGENTS.md', content: BODY },
  { file: '.agents/rules/vibeshield.md', content: BODY },
  { file: '.cursor/rules/vibeshield.mdc', content: mdcFrontmatter + BODY },
  { file: '.cursorrules', content: BODY },
  { file: '.windsurf/rules/vibeshield.md', content: BODY },
  { file: '.windsurfrules', content: BODY },
  { file: '.github/copilot-instructions.md', content: BODY },
];

const check = process.argv.includes('--check');
let drift = 0;

for (const { file, content } of targets) {
  const path = join(repoRoot, file);
  if (check) {
    let existing = null;
    try {
      existing = readFileSync(path, 'utf8');
    } catch {
      existing = null;
    }
    if (existing !== content) {
      console.error(`drift: ${file} is out of sync with AgentRulesBody`);
      drift++;
    }
    continue;
  }
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, content);
  console.log(`ok   ${file}`);
}

if (check) {
  console.log(drift ? `${drift} agent file(s) out of sync` : `agent rules in sync (${targets.length} files)`);
  process.exit(drift ? 1 : 0);
}
console.log(`\nsynced ${targets.length} agent rule files from AgentRulesBody`);
