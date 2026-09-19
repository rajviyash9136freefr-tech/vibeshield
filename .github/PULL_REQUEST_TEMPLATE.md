## What does this PR do?

<!-- One paragraph. What changed, and why. Link the issue it closes, if any. -->

Closes #

## Type of change

- [ ] Bug fix (a rule fires when it should not, or does not fire when it should)
- [ ] New rule (adds a failure mode to the pack)
- [ ] New command or flag
- [ ] Documentation
- [ ] Build, CI or release tooling
- [ ] Something else

## The gates

Every one of these must pass. They are the same checks CI runs.

- [ ] `cd scanner && go test ./... && go vet ./...`
- [ ] `node scripts/sync-rules.mjs --check` — rule packs in sync
- [ ] `node scripts/sync-agent-rules.mjs --check` — generated agent files in sync
- [ ] `node scripts/audit-contract.mjs scanner/vibeshield` — 0 documented-but-missing surfaces
- [ ] `gofmt -l .` reports nothing for the files I touched

## If this adds a rule

- [ ] The rule is in `rules/core/` as YAML, and I ran `node scripts/sync-rules.mjs`
- [ ] It has a `message` explaining why this matters **for AI-generated code**
- [ ] It has a one-line `fix`, imperative voice
- [ ] I added a fixture under `fixtures/golden/` that it fires on
- [ ] I checked it does **not** fire on `fixtures/golden/js-clean`
- [ ] `vibeshield rules <NEW-ID>` shows it as active (not `RESERVED`)

## If this adds a command or flag

A new surface has to be reflected in four places, and the build fails if the
first two drift apart:

- [ ] `scanner/cmd/vibeshield/help.go` — the `docs()` entry
- [ ] `scanner/cmd/vibeshield/completion.go` — the `completionFlags` map
- [ ] `scripts/audit-contract.mjs` — the `COMMANDS` list, if it is a new verb
- [ ] `site/src/content/docs/` and the README, if it is user-facing

## Anything the reviewer should know?

<!-- Trade-offs, things you tried and rejected, follow-up work, known gaps. -->

## Output

<!-- If this changes what a command prints, paste the before and after. -->
