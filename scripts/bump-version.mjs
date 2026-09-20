#!/usr/bin/env node
// Version bump with explicit, auditable per-file replacements.
//
//   node scripts/bump-version.mjs 2.0.0 2.0.1
//   node scripts/bump-version.mjs 2.0.0 2.1.0 --pack 2.1.0
//
// The rule-pack version is deliberately separate: rules/core/*.yaml carries
// the PACK version, not the tool version, so a CLI-only release must leave it
// alone. Pass --pack only when the rules themselves changed.
//
// Replacements are listed per file rather than applied with a blanket regex,
// because fixture project manifests also contain "version": "1.0.0" and must
// never be touched.
import { readFileSync, writeFileSync, readdirSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const [from, to, ...rest] = process.argv.slice(2);
const packIdx = rest.indexOf('--pack');
const packTo = packIdx === -1 ? null : rest[packIdx + 1];

if (!from || !to) {
  console.error('usage: node scripts/bump-version.mjs <from> <to> [--pack <packVersion>]');
  process.exit(2);
}
for (const v of [from, to, ...(packTo ? [packTo] : [])]) {
  if (!/^\d+\.\d+\.\d+$/.test(v)) {
    console.error(`not a plain semver: ${v}`);
    process.exit(2);
  }
}

const repoRoot = fileURLToPath(new URL('..', import.meta.url));

const edits = [];
const add = (rel, f, t, expect = 1) => edits.push({ rel, from: f, to: t, expect });

// Rule packs — only when the rules themselves changed.
if (packTo) {
  for (const dir of ['rules/core', 'scanner/internal/rules/packs/core']) {
    for (const f of readdirSync(join(repoRoot, dir)).filter((n) => n.endsWith('.yaml'))) {
      add(`${dir}/${f}`, `\nversion: ${from}\n`, `\nversion: ${packTo}\n`);
    }
  }
}

// Tool version.
add('scanner/cmd/vibeshield/main.go', `var Version = "${from}"`, `var Version = "${to}"`);
add('npm/vibeshield/package.json', `"version": "${from}"`, `"version": "${to}"`);
add('npm/vibeshield/lib/run.js', `|| 'v${from}'`, `|| 'v${to}'`);
add('skill/.claude-plugin/plugin.json', `"version": "${from}"`, `"version": "${to}"`);
add('action/action.yml', `default: 'v${from}'`, `default: 'v${to}'`);
add('action/entrypoint.sh', `:-v${from}}`, `:-v${to}}`);
add('.pre-commit-hooks.yaml', `cmd/vibeshield@v${from}`, `cmd/vibeshield@v${to}`);
// The Action ref `vibeshield init` falls back to for a non-semver build.
add('scanner/internal/initcmd/init.go', `const defaultActionTag = "v${from}"`, `const defaultActionTag = "v${to}"`);
// The installers. These were missing from this list until v3, so both defaulted
// to v1.0.0 and the README's headline one-liner installed a two-major-versions-
// old release without saying so. A default that drifts silently is the worst
// kind of bug in an install path, so they are pinned here from now on.
add('scripts/install.sh', `VIBESHIELD_VERSION:-v${from}}`, `VIBESHIELD_VERSION:-v${to}}`);
add('scripts/install.ps1', `[string]$Version = "v${from}"`, `[string]$Version = "v${to}"`);
add('contracts/cli.md', `"version": "${from}",`, `"version": "${to}",`);
add('action/test/fixtures/scan-sample.json', `"version": "${from}",`, `"version": "${to}",`);

// Site copy + docs.
add('site/src/data/site.ts', `vibeshield/action@v${from}`, `vibeshield/action@v${to}`);
add('site/src/content/docs/cli.md', `"version": "${from}",`, `"version": "${to}",`);
add('site/src/content/docs/github-action.md', `action@v${from}`, `action@v${to}`, -1);
add('site/src/content/docs/github-action.md', `\`v${from}\``, `\`v${to}\``);
add('site/src/content/docs/pre-commit.md', `rev: v${from}`, `rev: v${to}`);
add('site/src/content/docs/quickstart.md', `action@v${from}`, `action@v${to}`);
add('site/src/content/docs/quickstart.md', `rev: v${from}`, `rev: v${to}`);
add('site/src/content/docs/installation.md', `cmd/vibeshield@v${from}`, `cmd/vibeshield@v${to}`);
add('site/src/content/docs/installation.md', `vibeshield ${from}`, `vibeshield ${to}`);
add('site/src/pages/install.astro', `action@v${from}`, `action@v${to}`);
add('site/src/pages/install.astro', `rev: v${from}`, `rev: v${to}`);
add('site/src/pages/install.astro', `cmd/vibeshield@v${from}`, `cmd/vibeshield@v${to}`);
add('README.md', `version-v${from}-blue`, `version-v${to}-blue`);
add('README.md', `/download/v${from}/`, `/download/v${to}/`, -1);
add('README.md', `cmd/vibeshield@v${from}`, `cmd/vibeshield@v${to}`, -1);

let failures = 0;
for (const { rel, from: f, to: t, expect } of edits) {
  const path = join(repoRoot, rel);
  if (!existsSync(path)) {
    console.error(`MISSING ${rel}`);
    failures++;
    continue;
  }
  const text = readFileSync(path, 'utf8');
  const count = text.split(f).length - 1;
  // expect = -1 means "one or more".
  if (expect === -1 ? count < 1 : count !== expect) {
    console.error(`SKIP ${rel}: found ${count} of ${JSON.stringify(f)}, expected ${expect === -1 ? '>=1' : expect}`);
    failures++;
    continue;
  }
  writeFileSync(path, text.split(f).join(t));
  console.log(`ok   ${rel}  (${count}×)`);
}

console.log(failures
  ? `\n${failures} file(s) need attention`
  : `\n${from} -> ${to} applied${packTo ? ` (pack -> ${packTo})` : ' (pack untouched)'}`);
process.exit(failures ? 1 : 0);
