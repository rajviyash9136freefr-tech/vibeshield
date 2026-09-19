# Security policy

## Reporting a vulnerability in VibeShield

**Do not open a public issue for a security problem.** Use GitHub's private
advisory flow instead:

**→ [Open a private security advisory](https://github.com/rajviyash9136freefr-tech/vibeshield/security/advisories/new)**

That keeps the details private until there is a fix. If you cannot use GitHub
advisories, open a normal issue that says only "I have a security report and
need a private channel" — with no details — and a maintainer will get in touch.

### What to include

- What the vulnerability is, and the impact you believe it has.
- The version you tested (`vibeshield version` prints it).
- A minimal reproduction: a repository, a file, or a command line.
- Whether you have told anyone else, and whether you plan to.

### What to expect

| Stage | Target |
|:---|:---|
| Acknowledgement | within 48 hours |
| Initial assessment | within 5 days |
| Fix for high or critical | within 14 days |
| Public disclosure | coordinated with you, after a fix ships |

We credit reporters in the advisory and the release notes unless you ask us not
to. We will not take legal action against anyone who reports in good faith and
gives us the window above.

## What counts as a vulnerability here

VibeShield is unusual: it is a security tool with **no server, no account and no
network code**. The interesting classes of bug are therefore narrower than for
a hosted product.

**In scope**

- A **false negative** where a rule claims to detect something and does not —
  particularly anything that would let a secret or a slopsquatted package pass
  the pre-commit hook or the PR gate unnoticed.
- **Secret leakage through output.** Findings are supposed to redact the matched
  region to four characters on each side. Anything that prints, logs, posts or
  writes a full credential to `vibeshield-fixes.log`, a PR comment, JSON, or
  SARIF is a real finding.
- **Rule-pack injection.** Packs are data, not code. A malformed pack must be a
  startup error (exit 2), never something the scanner evaluates. A YAML file
  that achieves code execution is critical.
- **Path traversal or writes outside the target.** `fix` writes only the files
  it reports, and only within the scanned tree. `init` writes at most three
  known paths.
- **Supply-chain issues in how we ship.** The installers and the npm launcher
  verify release assets against `sha256sums.txt`; anything that lets a
  tampered asset execute is critical.
- **Prompt injection in the agent skill.** The skill reads repository content —
  READMEs, issue text, `AGENTS.md` — and is supposed to treat all of it as
  untrusted data. Content that redirects the agent into an unsafe action is a
  vulnerability, not a feature gap.

**Out of scope**

- A rule that produces a **false positive**. That is a bug, and a normal issue
  is the right place for it.
- Anything that requires an attacker to already control the machine the scanner
  runs on, or the repository being scanned.
- The `--online` flag. It is reserved and makes no network calls in this build.
- Missing hardening headers on the marketing site. It is a static GitHub Pages
  build with no cookies, no accounts and no user input.

## How we handle your report

1. We reproduce it and decide whether it is in scope.
2. We write a fix with a test that fails before it and passes after.
3. We ship it in a patch release and publish a GitHub Security Advisory.
4. We credit you, unless you would rather we did not.

## Supported versions

The latest minor release receives security fixes. Because the tool ships as a
static binary with no server component, fixes are simple to adopt: update the
binary, or bump the pinned Action ref in your workflow.

| Version | Supported |
|:---|:---|
| 3.x | ✅ |
| 2.x | security fixes only |
| 1.x | ❌ |

## Verifying a release

Every release publishes `sha256sums.txt` alongside the platform archives. The
install scripts and the npm launcher verify against it automatically; if you
install by hand, check it yourself:

```bash
sha256sum -c sha256sums.txt --ignore-missing
```

A mismatch means the download is not what we published. Do not run it.
