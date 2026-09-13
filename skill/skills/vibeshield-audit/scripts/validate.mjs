#!/usr/bin/env node
/**
 * VibeShield findings validator — checks a vibeshield-findings.json array
 * against contracts/finding/schema.json (draft 2020-12, the subset that
 * matters) with zero dependencies. Exit 0 = valid, 1 = violations printed
 * one per line, 2 = usage/IO error. Ships inside the plugin so the skill's
 * Phase 3 works on any machine with just Node.
 */
import { readFileSync } from 'node:fs';

const file = process.argv[2];
if (!file) {
  console.error('usage: node validate.mjs <findings.json>');
  process.exit(2);
}

const SEVERITIES = ['critical', 'high', 'medium', 'low', 'info'];
const CATEGORIES = [
  'hallucinated-package', 'hardcoded-secret', 'insecure-api', 'license-missing',
  'insecure-default', 'dependency-risk', 'prompt-injection',
];
const ORIGINS = ['confirmed', 'likely', 'unknown'];
const RULE_ID = /^VS-(PKG|SEC|LIC|DEP|INJ)-\d{3}$/;
const ALLOWED = new Set([
  'schema_version', 'rule_id', 'severity', 'category', 'title', 'message',
  'file', 'line', 'end_line', 'column', 'snippet', 'fix', 'ai_origin',
  'confidence', 'dismissable', 'dismiss_hash', 'metadata',
]);
const REQUIRED = ['schema_version', 'rule_id', 'severity', 'category', 'title', 'message', 'file', 'ai_origin'];

let data;
try {
  data = JSON.parse(readFileSync(file, 'utf8'));
} catch (e) {
  console.error(`${file}: cannot read/parse: ${e.message}`);
  process.exit(2);
}

const errs = [];
const add = (i, msg) => errs.push(`[${file}] finding ${i}: ${msg}`);

if (!Array.isArray(data)) {
  console.error(`${file}: top level must be an array of findings`);
  process.exit(1);
}

data.forEach((f, i) => {
  if (typeof f !== 'object' || f === null || Array.isArray(f)) return add(i, 'must be an object');
  for (const k of REQUIRED) if (f[k] === undefined) add(i, `missing required "${k}"`);
  for (const k of Object.keys(f)) if (!ALLOWED.has(k)) add(i, `unknown property "${k}" (additionalProperties: false)`);
  if (f.schema_version !== 1) add(i, 'schema_version must be 1');
  if (f.rule_id !== undefined && !RULE_ID.test(f.rule_id)) add(i, `rule_id "${f.rule_id}" must match VS-(PKG|SEC|LIC|DEP|INJ)-###`);
  if (f.severity !== undefined && !SEVERITIES.includes(f.severity)) add(i, `severity "${f.severity}" invalid`);
  if (f.category !== undefined && !CATEGORIES.includes(f.category)) add(i, `category "${f.category}" invalid`);
  if (f.ai_origin !== undefined && !ORIGINS.includes(f.ai_origin)) add(i, `ai_origin "${f.ai_origin}" invalid`);
  if (f.title !== undefined && (typeof f.title !== 'string' || f.title.length === 0)) add(i, 'title must be a non-empty string');
  if (typeof f.title === 'string' && [...f.title].length > 80) add(i, 'title exceeds 80 chars');
  if (f.message !== undefined && (typeof f.message !== 'string' || f.message.length === 0)) add(i, 'message must be non-empty');
  if (f.file !== undefined && (typeof f.file !== 'string' || f.file.length === 0)) add(i, 'file must be non-empty');
  if (typeof f.file === 'string' && (f.file.includes('\\') || f.file.startsWith('/') || f.file.startsWith('./'))) {
    add(i, `file "${f.file}" must be repo-relative POSIX (no \\, no leading / or ./)`);
  }
  for (const k of ['line', 'end_line', 'column']) {
    if (f[k] !== undefined && (!Number.isInteger(f[k]) || f[k] < 1)) add(i, `${k} must be an integer ≥ 1`);
  }
  if (f.confidence !== undefined && (typeof f.confidence !== 'number' || f.confidence < 0 || f.confidence > 1)) {
    add(i, 'confidence must be a number in [0,1]');
  }
  if (f.dismissable !== undefined && typeof f.dismissable !== 'boolean') add(i, 'dismissable must be boolean');
  if (f.metadata !== undefined && (typeof f.metadata !== 'object' || f.metadata === null || Array.isArray(f.metadata))) {
    add(i, 'metadata must be an object');
  }
  // Redaction law: no unredacted-looking secret material in snippets.
  if (typeof f.snippet === 'string' && /(sk|pk)-(live|test|proj)-[A-Za-z0-9]{20,}|ghp_[A-Za-z0-9]{36}|AKIA[0-9A-Z]{16}/.test(f.snippet)) {
    add(i, 'snippet appears to contain an unredacted secret (keep 4+4, mask middle with •)');
  }
});

// Cross-finding dedupe on dismiss_hash.
const seen = new Map();
data.forEach((f, i) => {
  if (f && typeof f.dismiss_hash === 'string' && f.dismiss_hash) {
    if (seen.has(f.dismiss_hash)) add(i, `duplicate dismiss_hash ${f.dismiss_hash} (also finding ${seen.get(f.dismiss_hash)})`);
    else seen.set(f.dismiss_hash, i);
  }
});

if (errs.length) {
  for (const e of errs) console.error(e);
  console.error(`${file}: ${errs.length} violation(s)`);
  process.exit(1);
}
console.log(`${file}: OK — ${data.length} finding(s)`);
