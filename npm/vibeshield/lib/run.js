// Fetch-on-first-run launcher for the vibeshield Go binary.
// Plain Node (>=18), zero dependencies, Windows-first (Git Bash dev box).
//
// Release assets follow action/entrypoint.sh's naming law:
//   vibeshield-{os}-{arch}.tar.gz | .zip  (os: linux|darwin|windows,
//                                          arch: x86_64|aarch64)
// plus sha256sums.txt covering them. We verify the checksum before exec.
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import {
  existsSync, mkdirSync, readFileSync, writeFileSync,
  renameSync, chmodSync, rmSync, readdirSync, statSync,
} from 'node:fs';
import { tmpdir, homedir } from 'node:os';
import { join } from 'node:path';

const REPO = 'rajviyash9136freefr-tech/vibeshield';
const VERSION = process.env.VIBESHIELD_VERSION || 'v2.0.1';

function cacheDir() {
  if (process.env.VIBESHIELD_HOME) return process.env.VIBESHIELD_HOME;
  const base =
    process.platform === 'win32'
      ? join(process.env.LOCALAPPDATA || join(homedir(), 'AppData', 'Local'), 'vibeshield')
      : join(process.env.XDG_CACHE_HOME || join(homedir(), '.cache'), 'vibeshield');
  return join(base, VERSION);
}

function assetBase() {
  const os = { linux: 'linux', darwin: 'darwin', win32: 'windows' }[process.platform];
  const arch = { x64: 'x86_64', arm64: 'aarch64' }[process.arch];
  if (!os || !arch) {
    fail(`unsupported platform ${process.platform}/${process.arch} — grab a binary from https://github.com/${REPO}/releases or build: go install github.com/${REPO}/scanner/cmd/vibeshield@latest`);
  }
  return { name: `vibeshield-${os}-${arch}`, ext: os === 'windows' ? 'zip' : 'tar.gz', bin: os === 'windows' ? 'vibeshield.exe' : 'vibeshield' };
}

function fail(msg) {
  console.error(`vibeshield: ${msg}`);
  process.exit(2);
}

async function get(url, buf = true) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) throw new Error(`${url} → HTTP ${res.status}`);
  return buf ? Buffer.from(await res.arrayBuffer()) : res.text();
}

function sha256(buf) {
  return createHash('sha256').update(buf).digest('hex');
}

async function downloadBinary(dest) {
  const { name, ext } = assetBase();
  const base = `https://github.com/${REPO}/releases/download/${VERSION}`;
  const tmp = join(tmpdir(), `${name}-${Date.now()}.${ext}`);
  let archive;
  try {
    process.stderr.write(`vibeshield ${VERSION}: first run — downloading the scanner (~6 MB, one time)…\n`);
    archive = await get(`${base}/${name}.${ext}`);
    const sums = await get(`${base}/sha256sums.txt`, false);
    const want = Object.fromEntries(
      sums.split('\n').map((l) => l.trim().split(/\s+/)).filter((p) => p.length >= 2).map((p) => [p[1].replace(/^\*/, ''), p[0]]),
    );
    const got = sha256(archive);
    if (want[`${name}.${ext}`] && want[`${name}.${ext}`] !== got) {
      fail(`checksum mismatch for ${name}.${ext} — refusing to run`);
    }
  } catch (err) {
    fail(`could not download the scanner (${err.message}).\n` +
      `  Offline? Build it once: go install github.com/${REPO}/scanner/cmd/vibeshield@latest\n` +
      `  Then point the wrapper at it: VIBESHIELD_BIN=/path/to/vibeshield npx vibeshield …`);
  }
  writeFileSync(tmp, archive);

  // Extract. tar is preinstalled on Windows 10+ and every Git-Bash box;
  // unzip likewise.
  const dir = join(tmpdir(), `vibeshield-x-${Date.now()}`);
  mkdirSync(dir, { recursive: true });
  const tool = ext === 'zip' ? 'unzip' : 'tar';
  const args = ext === 'zip' ? ['-o', tmp, '-d', dir] : ['-xzf', tmp, '-C', dir];
  const r = spawnSync(tool, args, { stdio: 'inherit' });
  rmSync(tmp, { force: true });
  if (r.status !== 0) fail(`extraction failed (is ${tool} available?)`);
  const bin = findDeep(dir, assetBase().bin);
  if (!bin) fail('archive did not contain the vibeshield binary');
  mkdirSync(cacheDir(), { recursive: true });
  renameSync(bin, dest);
  chmodSync(dest, 0o755);
  rmSync(dir, { recursive: true, force: true });
}

function findDeep(dir, file) {
  for (const e of readdirSync(dir)) {
    const p = join(dir, e);
    if (statSync(p).isDirectory()) {
      const hit = findDeep(p, file);
      if (hit) return hit;
    } else if (e === file) {
      return p;
    }
  }
  return null;
}

export async function main(argv) {
  const { bin } = assetBase();
  const dest = join(cacheDir(), bin);
  const run = (path) => {
    const r = spawnSync(path, argv, { stdio: 'inherit' });
    if (r.error) fail(r.error.message);
    process.exit(r.status ?? 1);
  };

  // No download when the user supplied their own binary.
  if (process.env.VIBESHIELD_BIN) {
    if (!existsSync(process.env.VIBESHIELD_BIN)) fail(`VIBESHIELD_BIN=${process.env.VIBESHIELD_BIN} not found`);
    return run(process.env.VIBESHIELD_BIN);
  }
  if (existsSync(dest)) return run(dest);

  await downloadBinary(dest);
  run(dest);
}
