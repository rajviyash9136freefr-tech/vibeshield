#!/usr/bin/env node
// Contract audit: every command, flag and format value the docs promise, run
// against the built binary. This is how `vibeshield init` and `--format sarif`
// were found documented-but-missing.
//
//   cd scanner && go build -o vibeshield.exe ./cmd/vibeshield
//   node scripts/audit-contract.mjs
//
// Exits non-zero when a documented surface is missing, so CI can gate on it.
import { readFileSync, readdirSync, existsSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = fileURLToPath(new URL('..', import.meta.url));
const isWin = process.platform === 'win32';
const bin = process.argv[2] || join(repoRoot, 'scanner', isWin ? 'vibeshield.exe' : 'vibeshield');

if (!existsSync(bin)) {
  console.error(`no binary at ${bin}`);
  console.error(`build it first: cd scanner && go build -o vibeshield${isWin ? '.exe' : ''} ./cmd/vibeshield`);
  process.exit(2);
}

// Flags that belong to OTHER tools mentioned in the same docs. Each needs a
// reason, so this list cannot quietly become a dumping ground.
const NOT_OURS = new Map([
  ['--check', 'sync-rules.mjs / sync-agent-rules.mjs (Node scripts)'],
  ['--cached', 'git diff --cached'],
  ['--no-verify', 'git commit --no-verify'],
  ['--quick', 'the Claude Code skill: /vibeshield:vibeshield-audit --quick'],
  ['--reason', 'the Action PR-comment command: /vibeshield accept <id> --reason'],
]);

// `vibeshield accept ...` is an Action PR-comment command, not a CLI verb.
const NOT_A_COMMAND = new Set(['accept']);

// --- what the docs claim ----------------------------------------------------

const docFiles = [
  'contracts/cli.md',
  'contracts/rulepack.md',
  'README.md',
  'docs/agents.md',
  ...readdirSync(join(repoRoot, 'site/src/content/docs')).map((f) => `site/src/content/docs/${f}`),
].filter((f) => existsSync(join(repoRoot, f)));

const text = docFiles.map((f) => readFileSync(join(repoRoot, f), 'utf8')).join('\n');

const commands = new Set();
// Require "vibeshield" NOT to be preceded by "/" — that form is the Action's
// PR-comment command namespace, not a CLI verb.
for (const m of text.matchAll(/(?<![/\w])vibeshield ([a-z][a-z-]{1,20})\b/g)) {
  if (!NOT_A_COMMAND.has(m[1])) commands.add(m[1]);
}

// Only lines that DEFINE a flag, i.e. "  --flag ..." or "  -v, --verbose ...".
const flags = new Set();
for (const line of text.split('\n')) {
  const m = line.match(/^\s*(?:-[A-Za-z],\s*)?(--[a-z][a-z-]{1,20})\b/);
  if (m) flags.add(m[1]);
}

const COMMANDS = ['scan', 'fix', 'init', 'version', 'search', 'agents', 'ui', 'menu', 'console', 'help', 'find', 'agent'];

const formats = new Set();
for (const m of text.matchAll(/pretty\s*\|\s*json\s*\|\s*([a-z| ]+)/g)) {
  for (const f of m[1].split('|').map((s) => s.trim()).filter((s) => /^[a-z]+$/.test(s))) formats.add(f);
}

// --- what the binary does ---------------------------------------------------

const run = (args) => spawnSync(bin, args, { encoding: 'utf8', cwd: repoRoot });
const SUBCOMMANDS = ['scan', 'fix', 'init', 'search', 'agents'];

// A path that never exists, so a probe that gets past flag parsing fails fast
// on the path check instead of actually scanning the repository.
const DEAD_PATH = join(repoRoot, '.__vibeshield_audit_no_such_path__');

// Go's flag package rejects an unknown flag with this exact wording, and does
// it before any work happens. Probing beats grepping the help text: every
// subcommand prints the *global* usage block, so a text search would claim
// `scan` accepts `--no-hook`, which it does not.
const UNKNOWN = /flag provided but not defined/;

function acceptsFlag(cmd, flag) {
  const r = run([cmd, flag, DEAD_PATH]);
  const out = (r.stdout || '') + (r.stderr || '');
  return !UNKNOWN.test(out);
}

const flagSources = (flag) => SUBCOMMANDS.filter((c) => acceptsFlag(c, flag));

let gaps = 0;
console.log(`binary: ${bin}\n`);

console.log('COMMANDS documented -> implemented');
for (const c of [...commands].sort()) {
  const known = COMMANDS.includes(c);
  if (!known) gaps++;
  console.log(`  ${(known ? 'ok' : 'MISSING').padEnd(8)} vibeshield ${c}`);
}

console.log('\nFLAGS documented -> accepted');
for (const f of [...flags].sort()) {
  if (NOT_OURS.has(f)) {
    console.log(`  n/a      ${f}  (${NOT_OURS.get(f)})`);
    continue;
  }
  const where = flagSources(f);
  if (where.length === 0) {
    gaps++;
    console.log(`  MISSING  ${f}  (no subcommand accepts it)`);
  } else {
    console.log(`  ok       ${f}  (${where.join(', ')})`);
  }
}

console.log('\nFORMATS documented -> implemented');
for (const f of [...formats].sort()) {
  const r = run(['scan', '.', '--format', f, '--no-color']);
  const out = (r.stdout || '') + (r.stderr || '');
  const missing = /lands in v|unknown --format|not in this build/.test(out);
  if (missing) gaps++;
  console.log(`  ${(missing ? 'MISSING' : 'ok').padEnd(8)} --format ${f}`);
}

console.log(`\n${gaps} documented-but-missing surface(s)`);
process.exit(gaps ? 1 : 0);
