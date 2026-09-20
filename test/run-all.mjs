#!/usr/bin/env node
// test/run-all.mjs — Comprehensive test runner for VibeShield CLI & Scanner
import { spawnSync } from 'node:child_process';
import { readFileSync, existsSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const rootDir = fileURLToPath(new URL('..', import.meta.url));
const isWin = process.platform === 'win32';
const binPath = resolve(rootDir, 'scanner', isWin ? 'vibeshield.exe' : 'vibeshield');
const pkg = JSON.parse(readFileSync(join(rootDir, 'npm', 'vibeshield', 'package.json'), 'utf8'));
const currentVersion = pkg.version;

if (!existsSync(binPath)) {
  console.error(`Error: built binary not found at ${binPath}`);
  console.error(`Build it first: cd scanner && go build -o ${isWin ? 'vibeshield.exe' : 'vibeshield'} ./cmd/vibeshield`);
  process.exit(1);
}

let passed = 0;
let failed = 0;
const results = [];

function runBin(args, options = {}) {
  const env = { ...process.env, ...options.env };
  const res = spawnSync(binPath, args, {
    cwd: options.cwd || rootDir,
    encoding: 'utf8',
    env,
    input: options.input,
  });
  return {
    code: res.status ?? (res.error ? 1 : 0),
    stdout: res.stdout || '',
    stderr: res.stderr || '',
    error: res.error,
  };
}

function test(name, fn) {
  try {
    fn();
    passed++;
    results.push({ name, status: 'PASS' });
    console.log(`  ✓ PASS: ${name}`);
  } catch (err) {
    failed++;
    results.push({ name, status: 'FAIL', error: err.message });
    console.error(`  ✗ FAIL: ${name}\n     ${err.message}`);
  }
}

function assert(condition, message) {
  if (!condition) throw new Error(message || 'Assertion failed');
}

console.log(`\n======================================================`);
console.log(`🛡️  VibeShield Comprehensive Test Suite`);
console.log(`Binary: ${binPath}`);
console.log(`======================================================\n`);

// -----------------------------------------------------------------------------
// 1. Fixture Scans & Detection Accuracy (Phase 1 & 2)
// -----------------------------------------------------------------------------
console.log(`[1/5] Testing Fixtures & Detection Rules...`);

const fixtures = [
  'vulnerable-node-app',
  'vulnerable-react-app',
  'vulnerable-supabase-app',
  'vulnerable-python-app',
  'clean-app',
];

for (const fixName of fixtures) {
  test(`Fixture scan: ${fixName}`, () => {
    const fixDir = join(rootDir, 'test', 'fixtures', fixName);
    const expectedFile = join(fixDir, 'expected-findings.json');
    const expected = JSON.parse(readFileSync(expectedFile, 'utf8'));

    const r = runBin(['scan', fixDir, '--format', 'json']);
    assert(r.code === 0 || r.code === 1, `Scan returned unexpected exit code ${r.code}`);

    let report;
    try {
      report = JSON.parse(r.stdout);
    } catch (e) {
      throw new Error(`Failed to parse JSON report: ${e.message}\nSTDOUT: ${r.stdout}\nSTDERR: ${r.stderr}`);
    }

    const findingCount = report.findings ? report.findings.length : 0;
    if (expected.min_findings !== undefined) {
      assert(findingCount >= expected.min_findings, `Expected at least ${expected.min_findings} findings, got ${findingCount}`);
    }
    if (expected.max_findings !== undefined) {
      assert(findingCount <= expected.max_findings, `Expected at most ${expected.max_findings} findings, got ${findingCount}`);
    }

    if (expected.required_rules) {
      const foundRules = new Set(report.findings.map(f => f.rule_id));
      for (const reqRule of expected.required_rules) {
        assert(foundRules.has(reqRule), `Expected finding with rule ${reqRule}, but it was not detected`);
      }
    }
  });
}

// -----------------------------------------------------------------------------
// 2. Edge Case Handling
// -----------------------------------------------------------------------------
console.log(`\n[2/5] Testing Edge Cases & Path Robustness...`);

test('Edge case: empty folder', () => {
  const p = join(rootDir, 'test', 'fixtures', 'edge-cases', 'empty-folder');
  const r = runBin(['scan', p, '--format', 'json']);
  assert(r.code === 0, `Expected exit 0 on empty folder, got ${r.code}`);
  const report = JSON.parse(r.stdout);
  assert(report.findings.length === 0, 'Expected 0 findings');
  assert(report.scan.files_scanned === 0, 'Expected 0 files scanned');
});

test('Edge case: path with spaces and unicode 🚀', () => {
  const p = join(rootDir, 'test', 'fixtures', 'edge-cases', 'path with spaces and unicode 🚀');
  const r = runBin(['scan', p, '--format', 'json']);
  assert(r.code === 0, `Expected exit 0 on unicode path, got ${r.code}`);
  const report = JSON.parse(r.stdout);
  assert(report.findings.length === 0, 'Expected 0 findings');
  assert(report.scan.files_scanned === 1, 'Expected 1 file scanned');
});

test('Edge case: binary file skipping and long lines', () => {
  const p = join(rootDir, 'test', 'fixtures', 'edge-cases', 'binary-and-long-lines');
  const r = runBin(['scan', p, '--format', 'json']);
  assert(r.code === 0, `Expected exit 0, got ${r.code}`);
  const report = JSON.parse(r.stdout);
  // Binary .png must be skipped; only .ts file scanned
  assert(report.scan.files_scanned === 1, `Expected 1 file scanned, got ${report.scan.files_scanned}`);
});

test('Edge case: nested monorepo structure', () => {
  const p = join(rootDir, 'test', 'fixtures', 'edge-cases', 'monorepo');
  const r = runBin(['scan', p, '--format', 'json']);
  assert(r.code === 1, `Expected exit 1 (critical finding in nested app), got ${r.code}`);
  const report = JSON.parse(r.stdout);
  assert(report.scan.files_scanned === 2, `Expected 2 files scanned, got ${report.scan.files_scanned}`);
  assert(report.findings.some(f => f.rule_id === 'VS-SEC-001'), 'Expected VS-SEC-001 in nested monorepo app');
});

// -----------------------------------------------------------------------------
// 3. Command Matrix & Flag Variations (Phase 2)
// -----------------------------------------------------------------------------
console.log(`\n[3/5] Testing Command Matrix & Flag Options...`);

test('Command: version', () => {
  const r = runBin(['version']);
  assert(r.code === 0, `Expected exit 0, got ${r.code}`);
  assert(r.stdout.includes(`vibeshield ${currentVersion}`), `Expected version ${currentVersion} in stdout, got: ${r.stdout}`);
  assert(r.stdout.includes('rule packs: core'), 'Expected rule packs line in output');
});

test('Command: help overview and command help', () => {
  const r1 = runBin(['help']);
  assert(r1.code === 0, 'help exited non-zero');
  assert(r1.stdout.includes('Everyday:'), 'help overview missing sections');

  const r2 = runBin(['help', 'scan']);
  assert(r2.code === 0, 'help scan exited non-zero');
  assert(r2.stdout.includes('vibeshield scan — audit a project'), 'help scan missing details');
});

test('Command: doctor (pretty and json)', () => {
  const r1 = runBin(['doctor', '.']);
  assert(r1.code === 0 || r1.code === 1, `doctor returned unexpected code ${r1.code}`);
  assert(r1.stdout.includes('project'), 'doctor output missing project header');

  const r2 = runBin(['doctor', '.', '--format', 'json']);
  assert(r2.code === 0 || r2.code === 1, `doctor json returned unexpected code ${r2.code}`);
  const doc = JSON.parse(r2.stdout);
  assert(doc.schema === 'vibeshield.doctor/v1', `Unexpected schema ${doc.schema}`);
  assert(Array.isArray(doc.checks), 'Expected checks array in doctor json');
});

test('Command: search and rules', () => {
  const r1 = runBin(['search', 'aws']);
  assert(r1.code === 0, 'search aws exited non-zero');
  assert(r1.stdout.includes('result(s) for "aws"'), 'search output missing query header');

  const r2 = runBin(['rules', 'VS-SEC-017', '--format', 'json']);
  assert(r2.code === 0, 'rules VS-SEC-017 exited non-zero');
  const res = JSON.parse(r2.stdout);
  assert(res.results && res.results.length > 0, 'Expected finding result for VS-SEC-017');
  assert(res.results[0].title.includes('OpenAI API key'), 'Unexpected rule title in search results');
});

test('Command: agents (--body and --markdown)', () => {
  const r1 = runBin(['agents', '--body']);
  assert(r1.code === 0, 'agents --body exited non-zero');
  assert(r1.stdout.includes('HALLUCINATED PACKAGES'), 'agents --body missing rule 1');
  assert(r1.stdout.includes('SECRETS'), 'agents --body missing rule 2');

  const r2 = runBin(['agents', '--markdown']);
  assert(r2.code === 0, 'agents --markdown exited non-zero');
  assert(r2.stdout.includes('| Agent | File | Scope |'), 'agents --markdown missing table header');
});

test('Command: completion', () => {
  for (const sh of ['bash', 'zsh', 'fish', 'powershell']) {
    const r = runBin(['completion', sh]);
    assert(r.code === 0, `completion ${sh} exited non-zero`);
    assert(r.stdout.length > 50, `completion ${sh} output too short`);
  }
});

test('Command: init (--dry-run)', () => {
  const r = runBin(['init', 'test/fixtures/clean-app', '--dry-run']);
  assert(r.code === 0, `init --dry-run exited non-zero: ${r.stderr}`);
  assert(r.stdout.includes('--dry-run: nothing was written'), 'Expected dry run notice');
});

test('Formats: sarif and github', () => {
  const r1 = runBin(['scan', 'test/fixtures/vulnerable-react-app', '--format', 'sarif']);
  assert(r1.code === 1, 'Expected exit 1 on sarif finding');
  const sarif = JSON.parse(r1.stdout);
  assert(sarif.version === '2.1.0', `Expected sarif 2.1.0, got ${sarif.version}`);
  assert(sarif.runs && sarif.runs.length > 0, 'Expected runs array in SARIF');

  const r2 = runBin(['scan', 'test/fixtures/vulnerable-react-app', '--format', 'github']);
  assert(r2.code === 1, 'Expected exit 1 on github format');
  assert(r2.stdout.includes('::error file=src/App.tsx'), 'Expected ::error github annotation');
});

test('Fix: --dry-run mode', () => {
  const r = runBin(['fix', 'test/fixtures/vulnerable-python-app', '--dry-run']);
  assert(r.code === 0, `fix --dry-run exited with code ${r.code}`);
  assert(r.stdout.includes('--dry-run: nothing was changed') || r.stdout.includes('patch(es)'), 'Expected dry-run preview in output');
});

// -----------------------------------------------------------------------------
// 4. Error Handling & Negative Tests (Phase 2 & 4)
// -----------------------------------------------------------------------------
console.log(`\n[4/5] Testing Error Handling, Typos & Guardrails...`);

test('Error: nonexistent path', () => {
  const r = runBin(['scan', '__nonexistent_path_xyz__']);
  assert(r.code === 2, `Expected exit 2 for nonexistent path, got ${r.code}`);
  assert(r.stderr.includes('not a readable directory'), `Expected readable directory error, got: ${r.stderr}`);
});

test('Error: unknown flag', () => {
  const r = runBin(['scan', '.', '--unrecognized-flag-xyz']);
  assert(r.code === 2, `Expected exit 2 for unknown flag, got ${r.code}`);
  assert(r.stderr.includes('flag provided but not defined'), `Expected flag error message, got: ${r.stderr}`);
});

test('Error: typo suggestion', () => {
  const r = runBin(['scna', '.']);
  assert(r.code === 2, `Expected exit 2 for typo command, got ${r.code}`);
  assert(r.stderr.includes('Did you mean `vibeshield scan`?'), `Expected typo suggestion, got: ${r.stderr}`);
});

test('Environment: NO_COLOR=1 removes ANSI escapes', () => {
  const r = runBin(['scan', 'test/fixtures/vulnerable-react-app'], { env: { NO_COLOR: '1' } });
  // ANSI escape regex: \x1b\[[0-9;]*m
  const hasAnsi = /\x1b\[[0-9;]*m/.test(r.stdout);
  assert(!hasAnsi, 'Expected zero ANSI color codes when NO_COLOR=1 is set');
});

// -----------------------------------------------------------------------------
// 5. NPM Package Launcher Integration (Phase 2)
// -----------------------------------------------------------------------------
console.log(`\n[5/5] Testing NPM Wrapper & Launcher...`);

test('NPM Launcher: executes via bin/vibeshield.js with VIBESHIELD_BIN', () => {
  const launcherPath = join(rootDir, 'npm', 'vibeshield', 'bin', 'vibeshield.js');
  const res = spawnSync(process.execPath, [launcherPath, 'version'], {
    cwd: rootDir,
    encoding: 'utf8',
    env: { ...process.env, VIBESHIELD_BIN: binPath },
  });
  assert(res.status === 0, `NPM launcher failed with exit code ${res.status}: ${res.stderr}`);
  assert(res.stdout.includes(`vibeshield ${currentVersion}`), `Expected version in launcher output, got: ${res.stdout}`);
});

// -----------------------------------------------------------------------------
// Summary Report
// -----------------------------------------------------------------------------
console.log(`\n======================================================`);
console.log(`Test Execution Summary:`);
console.log(`  Passed: ${passed}`);
console.log(`  Failed: ${failed}`);
console.log(`  Total:  ${passed + failed}`);
console.log(`======================================================\n`);

if (failed > 0) {
  process.exit(1);
}
