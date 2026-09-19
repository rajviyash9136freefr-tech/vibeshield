---
name: False positive or missed finding
about: A rule fired on code that is fine, or missed something it should have caught
title: '[RULE]: '
labels: ['bug', 'rules']
assignees: ''
---

<!--
Both directions of this are equally useful, and both are bugs. A rule that
fires on clean code gets the whole tool switched off; a rule that misses a real
problem is worse. Please say which one this is.
-->

### Direction

- [ ] **False positive** — the rule fired, but the code is fine
- [ ] **False negative** — the rule should have fired and did not

### Rule

<!-- The rule id from the finding, e.g. VS-SEC-017. Run `vibeshield rules <id>` to read it. -->

### Version

<!-- Paste the output of `vibeshield version`, and `vibeshield doctor` if the setup is unusual. -->

```console

```

### The code

<!--
The smallest snippet that shows the problem. Please redact real credentials —
replace them with something obviously fake like sk-proj-EXAMPLE.
-->

```
```

### The command you ran

```bash

```

### What you expected, and what happened

<!--
For a false positive: why is this code safe? What makes it a legitimate
pattern rather than something to fix?

For a false negative: what should the rule have matched, and why is that a
failure mode worth a rule?
-->

### Anything else

<!--
If you are proposing a rule change, note that rules are YAML in rules/core/ and
need no Go change — a PR is very welcome. See CONTRIBUTING.md.
-->
