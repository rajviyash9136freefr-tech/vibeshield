#!/usr/bin/env node
// Keeps the scanner's embedded core pack in sync with the versioned source
// of truth in rules/core/. Run before building/releasing the scanner; CI
// fails if the two ever drift.
//
// Windows-first (Git Bash) and plain Node — no platform-specific bits.
import { readFileSync, readdirSync, copyFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = fileURLToPath(new URL('..', import.meta.url));
const src = join(repoRoot, 'rules', 'core');
const dst = join(repoRoot, 'scanner', 'internal', 'rules', 'packs', 'core');

const files = readdirSync(src).filter((f) => f.endsWith('.yaml'));
const check = process.argv.includes('--check');

let drift = 0;
for (const f of files) {
  const a = readFileSync(join(src, f));
  if (check) {
    let b;
    try {
      b = readFileSync(join(dst, f));
    } catch {
      b = null;
    }
    if (!b || !a.equals(b)) {
      console.error(`drift: ${relative(repoRoot, join(dst, f))} differs from rules/core/${f}`);
      drift++;
    }
  } else {
    copyFileSync(join(src, f), join(dst, f));
  }
}

if (check) {
  console.log(drift ? `${drift} pack(s) out of sync` : `rules in sync (${files.length} packs)`);
  process.exit(drift ? 1 : 0);
} else {
  console.log(`synced ${files.length} core packs → scanner/internal/rules/packs/core`);
}
