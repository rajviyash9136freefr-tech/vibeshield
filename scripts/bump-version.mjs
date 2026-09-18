#!/usr/bin/env node
// One-shot v1.0.0 → v2.0.0 version bump. Explicit per-file replacements so
// fixture project versions (which are also "1.0.0") are never touched.
import { readFileSync, writeFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

const repoRoot = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1');

const edits = [];
const add = (rel, from, to, expect = 1) => edits.push({ rel, from, to, expect });

// Rule packs: the pack schema version, both copies of the source of truth.
for (const dir of ['rules/core', 'scanner/internal/rules/packs/core']) {
  for (const f of readdirSync(join(repoRoot, dir)).filter((f) => f.endsWith('.yaml'))) {
    add(`${dir}/${f}`, '\nversion: 1.0.0\n', '\nversion: 2.0.0\n');
  }
}

add('npm/vibeshield/package.json', '"version": "1.0.0"', '"version": "2.0.0"');
add('npm/vibeshield/lib/run.js', "|| 'v1.0.0'", "|| 'v2.0.0'");
add('skill/.claude-plugin/plugin.json', '"version": "1.0.0"', '"version": "2.0.0"');
add('action/action.yml', "default: 'v1.0.0'", "default: 'v2.0.0'");
add('.pre-commit-hooks.yaml', 'cmd/vibeshield@v1.0.0', 'cmd/vibeshield@v2.0.0');
add('contracts/cli.md', '"version": "1.0.0",', '"version": "2.0.0",');
add('contracts/rulepack.md', '\nversion: 1.0.0\n', '\nversion: 2.0.0\n');
add('action/test/fixtures/scan-sample.json', '"version": "1.0.0",', '"version": "2.0.0",');

// Site copy + docs.
add('site/src/data/site.ts', 'vibeshield/action@v1', 'vibeshield/action@v2');
add('site/src/content/docs/cli.md', '"version": "1.0.0",', '"version": "2.0.0",');
add('site/src/content/docs/github-action.md', 'action@v1', 'action@v2', 4);
add('site/src/content/docs/github-action.md', '`v1.0.0`', '`v2.0.0`');
add('site/src/content/docs/pre-commit.md', 'rev: v1.0.0', 'rev: v2.0.0');
add('site/src/content/docs/quickstart.md', 'action@v1', 'action@v2');
add('site/src/content/docs/quickstart.md', 'rev: v1.0.0', 'rev: v2.0.0');

let failures = 0;
for (const { rel, from, to, expect } of edits) {
  const path = join(repoRoot, rel);
  let text;
  try {
    text = readFileSync(path, 'utf8');
  } catch (err) {
    console.error(`MISSING ${rel}: ${err.message}`);
    failures++;
    continue;
  }
  const count = text.split(from).length - 1;
  if (count !== expect) {
    console.error(`SKIP ${rel}: found ${count} of ${JSON.stringify(from)}, expected ${expect}`);
    failures++;
    continue;
  }
  writeFileSync(path, text.split(from).join(to));
  console.log(`ok   ${rel}  (${count}×)`);
}
console.log(failures ? `\n${failures} file(s) need attention` : '\nall version bumps applied');
process.exit(failures ? 1 : 0);
