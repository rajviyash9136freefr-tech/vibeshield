# vibeshield (npm wrapper)

The `vibeshield` name on npm is a thin launcher for the Go scanner binary at
[github.com/rajviyash9136freefr-tech/vibeshield](https://github.com/rajviyash9136freefr-tech/vibeshield).

- **No install scripts.** `npm i -g vibeshield` (or `npx vibeshield …`) never
  touches the network at install time.
- **First run** downloads the right static binary for your OS/arch from GitHub
  Releases, verifies it against the release's `sha256sums.txt`, caches it under
  `~/.cache/vibeshield/<version>` (`%LOCALAPPDATA%\vibeshield` on Windows), and
  execs it. After that it's a single local process spawn — fully offline.
- **Pinning:** set `VIBESHIELD_VERSION` to grab a specific release; set
  `VIBESHIELD_BIN` to run a binary you built yourself (`go install
  github.com/rajviyash9136freefr-tech/vibeshield/scanner/cmd/vibeshield@latest`)
  and skip the download entirely.

Every flag and command is the scanner's — see `contracts/cli.md` in the repo
(`vibeshield scan .`, `--staged`, `--diff origin/main`, `--format json`, …).

MIT licensed, like everything here.
