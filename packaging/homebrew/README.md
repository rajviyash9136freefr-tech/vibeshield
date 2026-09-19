# Packaging

The scanner ships as a single static binary, so packaging is a matter of
describing a release asset to a package manager. Nothing in this directory is
built by CI; it exists to make the two distribution channels a one-command job
rather than a research task.

## Homebrew

```bash
# After the release is published:
node scripts/gen-homebrew-formula.mjs 3.0.0
```

That fetches the release's own `sha256sums.txt`, formats a formula, and writes
`packaging/homebrew/vibeshield.rb`. It refuses to run if the release is not
published or if any of the four macOS/Linux assets is missing a checksum —
which is the point: the formula can never carry a checksum for a tarball that
does not exist.

> The generated file is **not** committed, because a formula with placeholder
> checksums is worse than no formula. Generate it after tagging.

Then pick a channel:

| Channel | Effort | Result |
|:---|:---|:---|
| **Your own tap** | Create a repository named `homebrew-vibeshield`, put the file at `Formula/vibeshield.rb`, push. | `brew install rajviyash9136freefr-tech/vibeshield/vibeshield` works immediately. |
| **homebrew-core** | Open a PR at [Homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core) with the file at `Formula/v/vibeshield.rb`. | `brew install vibeshield`. Stricter bar: notability, a stable release history, and a clean `brew audit`. Read their CONTRIBUTING before opening it. |

Before submitting either way:

```bash
brew audit --new-formula packaging/homebrew/vibeshield.rb
```

The formula installs the binary and generates shell completions by calling
`vibeshield completion` — the same scripts `vibeshield completion` prints
anywhere else, so there is no second copy to maintain.

## npm

Handled by [`.github/workflows/publish-npm.yml`](../.github/workflows/publish-npm.yml)
on release, using OIDC trusted publishing. See the header of that workflow for
the one-time npmjs.com setup. No publish token is stored, deliberately.

## Adding a channel

If you add one, add it to:

1. The **Install** section of [`README.md`](../README.md).
2. [`site/src/content/docs/installation.md`](../site/src/content/docs/installation.md).
3. The `INSTALL_TABS` / `DOORS` arrays in the site, so the marketing page and the
   docs cannot disagree about how the tool is installed.

Do not document a channel that does not exist yet. This repository previously
advertised a `brew install vibeshield` that resolved to nothing, which is worse
than an incomplete install section: it costs the reader a failed command and
their trust in the rest of the page.
