# Growing VibeShield — a visibility playbook

The repository is not unpopular because the tool is bad. It is unpopular because
almost nobody has been given a reason to look, and the people who do look have
to work out what it is. This document is the plan, in priority order, with the
parts that are already done marked off.

It is written to be re-read: come back after a release, pick the next unchecked
box, and do it. Every item is either free or cheap, and none of them are
"growth hacks" — they are the things a project has to have before promotion
does anything except waste an afternoon.

---

## The one thing that matters most

**GitHub's own search ranks on repo name, description, topics, and
stars/forks/watchers. Google indexes the README.** Everything below is either
making the project findable by those two systems, or giving a person who lands
on it a reason to stay.

A repository that is hard to understand loses the visitor in about five
seconds. That is why the README rewrite, `vibeshield doctor`, and the docs site
came before any promotion in this plan: promotion spent on a confusing landing
page is promotion wasted.

---

## 1. Metadata — do this first, it is 10 minutes and it compounds

These are repository settings, not files, so they cannot be committed. Do them
in the browser.

- [ ] **About / description.** Settings → General → Description. GitHub search
      weights this heavily and Google shows it in the snippet. Use:

      > Offline security scanner for AI-generated code — hallucinated packages,
      > leaked secrets, insecure defaults. Go CLI, GitHub Action, pre-commit
      > hook. MIT.

      First word is the category, the platform is named, and it says what it
      finds rather than how great it is.

- [ ] **Topics.** Settings → Topics. Topics are how people filter GitHub search,
      so they are the highest-leverage field on the page. Use the list in the
      README's closing comment (it is kept in sync for exactly this reason):

      `vibe-coding`, `ai-coding`, `ai-agents`, `ai-security`, `claude-code`,
      `codex`, `cursor`, `antigravity`, `github-copilot`, `windsurf`,
      `agents-md`, `slopsquatting`, `hallucinated-packages`, `secret-scanning`,
      `supply-chain-security`, `pre-commit`, `github-action`, `devsecops`,
      `sast`, `static-analysis`, `cli`, `golang`, `developer-tools`,
      `security-tools`

      Twenty is the maximum. All twenty of those are terms people actually
      search for.

- [ ] **Website.** Settings → General → Website:
      `https://rajviyash9136freefr-tech.github.io/vibeshield/`

- [ ] **Social preview image.** Settings → General → Social preview. Upload
      **`site/public/social-preview.png`** — it is already generated at the
      required 1280×640, from `node scripts/gen-og.mjs` (or automatically on
      every `npm run build` in `site/`). Without it, every Slack, Discord, X and
      LinkedIn link to this repo renders as a grey rectangle with a tiny avatar.
      This is the single highest ratio of "effort" to "clicks" on this list, and
      it is now a two-click job.

- [ ] **Discussions.** Settings → Features → Discussions ✅. The README and the
      issue template config both link to it. It also gives "I have a question"
      somewhere to go that is not the issue tracker.

- [ ] **Pin the repository** on your GitHub profile, so the first thing anyone
      who clicks your name sees is the project.

---

## 2. The README is the landing page

Done in v3:

- [x] Opens with one sentence that says what it is and who it is for.
- [x] A real terminal transcript, not a description of one.
- [x] Install in one line, for every platform, above the fold.
- [x] A 60-second quickstart before the reference material.
- [x] Badges that are honest (version, stars, license) rather than decorative.
- [x] A star call-to-action in the Community section.
- [x] Keywords in the closing HTML comment, so GitHub's crawler and a human
      forking the repo both see them.

Still worth doing:

- [ ] **A GIF of the console.** The interactive console is the most
      distinctive thing in the project and a static transcript undersells it.
      Ten seconds of typing `cors` and watching the list filter is worth more
      than any paragraph. Record with `asciinema` and convert, or use a screen
      recorder; keep it under 2 MB.
- [ ] **Re-read it in six months** and delete anything that is no longer true.
      A README that lies is worse than a short one.

---

## 3. Releases are a distribution channel

A repository with no releases looks abandoned, and a release is a page that
Google indexes, that GitHub shows in the sidebar, and that people subscribe to.

- [x] `CHANGELOG.md` following Keep a Changelog, with a real v3.0.0 entry.
- [x] `release.yml` cross-compiles and publishes `sha256sums.txt` on every `v*`
      tag, and `generate_release_notes: true` is on.
- [ ] **Tag v3.0.0 and publish the release.** Nothing else in this document
      works until there is a release, because every install path points at one.
- [ ] **Write release notes a human would read.** The auto-generated list of
      commits is a starting point, not the notes. Lead with what a user can now
      do that they could not before. For v3 that is one sentence: "`vibeshield`
      now has a `doctor` command, per-command help, Tab completion, and a typo
      tells you what you meant."
- [ ] **Ship a patch release when you fix a bug.** Release cadence is a visible
      signal that the project is alive. Monthly is plenty; the goal is "not
      silent", not "not busy".

---

## 4. Get listed where people already look

Distribution beats polish. These are the places a developer looking for this
kind of tool actually goes, roughly in order of return on effort.

- [ ] **Awesome lists.** These are how a lot of people find dev tools, and they
      are permanent backlinks. Submit a PR to the relevant ones:
      - `awesome-cli-apps`, `awesome-go`, `awesome-security`,
        `awesome-static-analysis`, `awesome-devsecops`, `awesome-supply-chain`,
        `awesome-ai-coding`, `awesome-pre-commit`, `awesome-github-actions`
      - Read each list's contributing rules first. A rejected PR that ignored
        the format is a wasted hour; a good one is free traffic forever.
- [ ] **GitHub Action marketplace.** The Action is a composite action with a
      `branding` block already — publishing it to the Actions marketplace gives
      it a listing page and a search presence inside GitHub itself.
- [ ] **npm.** The publish workflow is written and waiting:
      [`.github/workflows/publish-npm.yml`](../.github/workflows/publish-npm.yml)
      publishes `npm/vibeshield` on release using OIDC trusted publishing, with
      provenance, and refuses to publish if `package.json` and `lib/run.js`
      disagree with the release tag. Publishing it makes `npx vibeshield scan .`
      work for everyone and puts the package in a second search index. The only
      remaining step is the one-time trusted-publisher setup on npmjs.com —
      described in the workflow header. **No token secret is needed, and none
      should be added.**
- [ ] **Homebrew.** `node scripts/gen-homebrew-formula.mjs 3.0.0` fetches the
      release's own `sha256sums.txt` and writes a ready-to-submit formula
      (see [`packaging/homebrew/README.md`](../packaging/homebrew/README.md)).
      A formula is the default install path for a large fraction of macOS
      developers, and `brew install vibeshield` is worth more than any tweet.

---

## 5. One good launch, not five bad ones

Pick a moment — the v3.0.0 release — and tell people once, well. Do not
drip-feed the same link into the same places.

- [ ] **Show HN.** Title it as the problem, not the product:
      *"Show HN: An offline scanner for the mistakes AI coding agents make"*.
      Link the repository, not the marketing site. Then **stay in the thread**
      and answer every question honestly, including the critical ones — the
      comments are the submission.
- [ ] **Reddit**, one post per community, tailored to each. `r/devops`,
      `r/netsec`, `r/ExperiencedDevs`, `r/programming`, `r/ClaudeAI`,
      `r/cursor`. Lead with the slopsquatting story and the offline promise;
      do not paste the same text four times, and do not post the same link to
      four subreddits on the same day.
- [ ] **A technical write-up** on your own blog or dev.to: *"The five ways AI
      coding agents get your supply chain wrong, and how to catch them"*. The
      tool is the last paragraph, not the first. This is the post that keeps
      working after the launch traffic dies.
- [ ] **Lobste.rs**, if the write-up is genuinely technical.
- [ ] **Do not** buy stars, join star-exchange groups, or post the same link
      daily. GitHub detects it, developers smell it, and it is the fastest way
      to make a real project look fake.

---

## 6. Make it easy to contribute — contributors are marketers

- [x] `CONTRIBUTING.md` with the exact gates CI runs.
- [x] `CODE_OF_CONDUCT.md`.
- [x] `SECURITY.md` with a scoped threat model — which matters more for a
      security tool than for anything else.
- [x] Issue templates: bug, rule request, false positive/missed finding, plus a
      `config.yml` that routes questions to Discussions and vulnerabilities to
      the private advisory flow.
- [x] `PULL_REQUEST_TEMPLATE.md` that asks the four questions a reviewer would
      otherwise have to ask.
- [ ] **Label a handful of issues `good first issue`** and keep them genuinely
      small. A first-time contributor who lands a PR tells people about it.
- [ ] **Respond to issues within 48 hours**, even if the answer is "I cannot
      get to this yet". An unanswered issue tells every future visitor the
      project is dead.
- [ ] **Credit contributors in the release notes.** It costs a line and it is
      the cheapest thing that makes someone come back.

---

## 7. Documentation is an SEO surface

The docs site is not only for users; it is the thing Google indexes and the
thing an LLM cites when someone asks an assistant "how do I scan AI-generated
code for leaked keys".

- [x] A real docs site with grouped navigation: getting started, the CLI,
      automation, reference.
- [x] Troubleshooting, installation and a v3 migration guide — the three pages
      that answer the three questions that generate issues.
- [x] Sitemap, `robots.txt`, JSON-LD (HowTo, FAQPage, SoftwareApplication,
      BreadcrumbList) so the pages are eligible for rich results.
- [ ] **Link the docs from places outside the repo** — the HN post, the
      write-up, an answer on Stack Overflow. GitHub Pages only ranks if
      something links to it.
- [ ] **Keep one canonical page per topic.** Duplicated CLI references in the
      README, the docs site and `contracts/cli.md` is exactly why
      `scripts/audit-contract.mjs` exists: it fails CI when the docs promise a
      flag the binary does not have.

---

## 8. What not to do

- **Do not claim what the tool cannot do.** This repository previously described
  a dashboard, a cloud API and a GitHub OAuth flow that do not exist. It also
  said `--online` made network calls, when the scanner links in no HTTP client
  at all. Anyone who read the code found the contradiction, and the first
  comment on a launch post would have been about that instead of the tool. The
  privacy, security and terms pages now describe what actually ships.
- **Do not hide the five reserved rules.** `vibeshield version` says
  `117 active · 5 reserved` and `vibeshield rules` marks them. Stating a
  limitation first is how a security tool earns trust; being caught
  overstating a rule count is how it loses it.
- **Do not optimise for stars.** Stars are a lagging indicator of people
  finding the tool useful. Optimise for "someone ran `vibeshield scan .` and it
  found something real", and the stars follow.

---

## The 30-day version

If you do nothing else, do these five, in order:

1. Set the About text, the 24 topics, and the social preview image.
2. Tag `v3.0.0` with release notes a human wrote.
3. Publish to npm and submit the Homebrew formula.
4. Open PRs to six awesome lists.
5. Write the slopsquatting post, publish it, then post it to Show HN and two
   subreddits — once each.
