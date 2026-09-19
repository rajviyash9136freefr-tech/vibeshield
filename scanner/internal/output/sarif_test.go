package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
)

// mk builds a finalized finding the way the scanner emits one.
func mk(ruleID, sev, cat, title, file string, line int, snippet string) *finding.Finding {
	f := finding.New(ruleID, sev, cat, title, "why it matters", file)
	f.Line = line
	f.Column = 3
	f.Snippet = snippet
	f.Fix = "do the safer thing"
	f.Finalize()
	return f
}

func reportOf(fs ...*finding.Finding) *scan.Report {
	rep := &scan.Report{
		SchemaVersion: 1,
		Tool:          "vibeshield",
		Version:       "2.0.0",
		Scan:          scan.ScanInfo{Mode: "full", FilesScanned: 12, DurationMs: 42},
	}
	for _, f := range fs {
		rep.Findings = append(rep.Findings, f)
		rep.Summary.Add(f.Severity)
	}
	return rep
}

// decode runs SARIF and unmarshals the result.
func decode(t *testing.T, rep *scan.Report) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if err := SARIF(&buf, rep); err != nil {
		t.Fatalf("SARIF returned an error: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("SARIF emitted invalid JSON: %v\n%s", err, buf.String())
	}
	return doc
}

func run0(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	runs, ok := doc["runs"].([]any)
	if !ok || len(runs) != 1 {
		t.Fatalf("expected exactly one run, got %v", doc["runs"])
	}
	return runs[0].(map[string]any)
}

func driverOf(t *testing.T, run map[string]any) map[string]any {
	t.Helper()
	return run["tool"].(map[string]any)["driver"].(map[string]any)
}

func TestSARIFEnvelope(t *testing.T) {
	doc := decode(t, reportOf())
	if doc["$schema"] != sarifSchema {
		t.Errorf("$schema = %v, want %v", doc["$schema"], sarifSchema)
	}
	if doc["version"] != "2.1.0" {
		t.Errorf("version = %v, want 2.1.0", doc["version"])
	}
	drv := driverOf(t, run0(t, doc))
	if drv["name"] != "VibeShield" {
		t.Errorf("driver name = %v", drv["name"])
	}
	if drv["version"] != "2.0.0" {
		t.Errorf("driver version = %v, want the report version", drv["version"])
	}
	if drv["informationUri"] == "" {
		t.Error("informationUri must be set — GitHub shows it on every alert")
	}
}

func TestSARIFEmptyReportUsesArraysNotNull(t *testing.T) {
	// GitHub rejects a run whose results or rules are null.
	drv := driverOf(t, run0(t, decode(t, reportOf())))
	if _, ok := drv["rules"].([]any); !ok {
		t.Fatalf("rules must be an array, got %T", drv["rules"])
	}
	run := run0(t, decode(t, reportOf()))
	if _, ok := run["results"].([]any); !ok {
		t.Fatalf("results must be an array, got %T", run["results"])
	}
}

func TestSARIFLevelMapping(t *testing.T) {
	cases := map[string]string{
		finding.SeverityCritical: "error",
		finding.SeverityHigh:     "error",
		finding.SeverityMedium:   "warning",
		finding.SeverityLow:      "note",
		finding.SeverityInfo:     "note",
	}
	for sev, want := range cases {
		if got := sarifLevel(sev); got != want {
			t.Errorf("sarifLevel(%q) = %q, want %q", sev, got, want)
		}
	}
}

func TestSARIFResultCarriesLevelAndFingerprint(t *testing.T) {
	f := mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "Key in source", "src/a.ts", 41, "const k = 'x'")
	run := run0(t, decode(t, reportOf(f)))
	results := run["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0].(map[string]any)
	if r["level"] != "error" {
		t.Errorf("high severity should be level error, got %v", r["level"])
	}
	if r["ruleId"] != "VS-SEC-017" {
		t.Errorf("ruleId = %v", r["ruleId"])
	}
	fp, ok := r["partialFingerprints"].(map[string]any)
	if !ok || fp[fingerprintKey] == "" {
		t.Fatalf("partialFingerprints must carry %s, got %v", fingerprintKey, r["partialFingerprints"])
	}
	if fp[fingerprintKey] != f.DismissHash {
		t.Errorf("fingerprint = %v, want the dismiss hash %v", fp[fingerprintKey], f.DismissHash)
	}
}

func TestSARIFRegionAndSnippet(t *testing.T) {
	f := mk("VS-SEC-001", finding.SeverityCritical, finding.CategoryHardcodedSecret, "AWS key", "src/aws.py", 7, "AKIAIOSFODNN7EXAMPLE")
	f.EndLine = 9
	r := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)
	pl := r["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
	if pl["artifactLocation"].(map[string]any)["uri"] != "src/aws.py" {
		t.Errorf("uri = %v, want the repo-relative path", pl["artifactLocation"])
	}
	region := pl["region"].(map[string]any)
	if region["startLine"].(float64) != 7 {
		t.Errorf("startLine = %v, want 7", region["startLine"])
	}
	if region["endLine"].(float64) != 9 {
		t.Errorf("endLine = %v, want 9", region["endLine"])
	}
	if region["snippet"].(map[string]any)["text"] == "" {
		t.Error("the snippet should be attached so the alert shows what matched")
	}
}

func TestSARIFColumnDefaultsToOne(t *testing.T) {
	// startColumn 0 is invalid SARIF; the region must never emit it.
	f := mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "t", "a.ts", 3, "x")
	f.Column = 0
	r := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)
	region := r["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)["region"].(map[string]any)
	if region["startColumn"].(float64) != 1 {
		t.Errorf("startColumn = %v, want 1", region["startColumn"])
	}
}

func TestSARIFFileLevelFindingOmitsRegion(t *testing.T) {
	// A manifest-level finding has no line. Emitting startLine 0 would be
	// invalid SARIF, so there must be no region at all.
	f := mk("VS-DEP-001", finding.SeverityMedium, finding.CategoryDependencyRisk, "unpinned", "requirements.txt", 0, "")
	r := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)
	pl := r["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
	if _, present := pl["region"]; present {
		t.Errorf("a line-less finding must not carry a region, got %v", pl["region"])
	}
	if pl["artifactLocation"].(map[string]any)["uri"] != "requirements.txt" {
		t.Error("the file location must still be reported")
	}
}

func TestSARIFRulesAreDedupedSortedAndIndexed(t *testing.T) {
	rep := reportOf(
		mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "b", "b.ts", 1, "x"),
		mk("VS-SEC-001", finding.SeverityCritical, finding.CategoryHardcodedSecret, "a", "a.ts", 2, "y"),
		mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "b", "c.ts", 3, "z"),
	)
	run := run0(t, decode(t, rep))
	rules := driverOf(t, run)["rules"].([]any)
	if len(rules) != 2 {
		t.Fatalf("a rule that fires twice must be listed once, got %d entries", len(rules))
	}
	ids := []string{}
	index := map[string]int{}
	for i, r := range rules {
		id := r.(map[string]any)["id"].(string)
		ids = append(ids, id)
		index[id] = i
	}
	if ids[0] != "VS-SEC-001" || ids[1] != "VS-SEC-017" {
		t.Errorf("rules must be sorted by id for byte-stable output, got %v", ids)
	}
	for _, r := range run["results"].([]any) {
		res := r.(map[string]any)
		want := index[res["ruleId"].(string)]
		if int(res["ruleIndex"].(float64)) != want {
			t.Errorf("ruleIndex for %v = %v, want %d", res["ruleId"], res["ruleIndex"], want)
		}
	}
}

func TestSARIFRuleHelpCarriesTheFix(t *testing.T) {
	f := mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "t", "a.ts", 1, "x")
	rules := driverOf(t, run0(t, decode(t, reportOf(f))))["rules"].([]any)
	rule := rules[0].(map[string]any)
	if rule["name"] != "VS-SEC-017" {
		t.Errorf("name = %v, want the whitespace-free rule id", rule["name"])
	}
	help := rule["help"].(map[string]any)
	if !strings.Contains(help["markdown"].(string), "do the safer thing") {
		t.Errorf("help should carry the one-line fix, got %v", help["markdown"])
	}
	if rule["defaultConfiguration"].(map[string]any)["level"] != "error" {
		t.Error("defaultConfiguration.level must match the rule severity")
	}
	props := rule["properties"].(map[string]any)
	if props["category"] != finding.CategoryHardcodedSecret {
		t.Errorf("rule properties should carry the category, got %v", props)
	}
}

func TestSARIFResultPropertiesCarryContractFields(t *testing.T) {
	f := mk("VS-PKG-001", finding.SeverityCritical, finding.CategoryHallucinatedPackage, "t", "package.json", 8, "x")
	f.AIOrigin = finding.OriginConfirmed
	f.Confidence = 0.85
	f.Metadata = map[string]any{"package": "fast-parse-utils-v3", "age_days": 9}

	props := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)["properties"].(map[string]any)
	for _, k := range []string{"severity", "category", "aiOrigin", "fix", "confidence", "dismissHash"} {
		if _, ok := props[k]; !ok {
			t.Errorf("result properties missing %q", k)
		}
	}
	// Contract metadata is namespaced so it cannot collide with our own keys.
	if props["vibeshield.package"] != "fast-parse-utils-v3" {
		t.Errorf("contract metadata should be namespaced, got %v", props)
	}
	if _, leaked := props["package"]; leaked {
		t.Error("contract metadata must not appear un-namespaced")
	}
}

func TestSARIFNeverEmitsAbsoluteOrBackslashPaths(t *testing.T) {
	// GitHub matches artifactLocation.uri against the repository tree, so an
	// absolute or Windows-style path silently produces an alert with no file.
	for _, in := range []string{`C:\Users\dev\proj\src\a.ts`, "/home/dev/src/a.ts", "./src/a.ts", "src\\a.ts"} {
		f := mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "t", in, 1, "x")
		uri := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)["artifactLocation"].(map[string]any)["uri"].(string)
		if strings.Contains(uri, `\`) {
			t.Errorf("input %q produced a backslash uri %q", in, uri)
		}
		if strings.HasPrefix(uri, "/") || strings.Contains(uri, ":") {
			t.Errorf("input %q produced a non-relative uri %q", in, uri)
		}
	}
}

func TestSARIFIsDeterministic(t *testing.T) {
	rep := reportOf(
		mk("VS-SEC-017", finding.SeverityHigh, finding.CategoryHardcodedSecret, "b", "b.ts", 1, "x"),
		mk("VS-SEC-001", finding.SeverityCritical, finding.CategoryHardcodedSecret, "a", "a.ts", 2, "y"),
		mk("VS-DEP-001", finding.SeverityLow, finding.CategoryDependencyRisk, "c", "requirements.txt", 0, ""),
	)
	var first bytes.Buffer
	if err := SARIF(&first, rep); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		var again bytes.Buffer
		if err := SARIF(&again, rep); err != nil {
			t.Fatal(err)
		}
		if again.String() != first.String() {
			t.Fatalf("run %d differs — output must be byte-stable for CI diffs", i)
		}
	}
}

func TestSARIFMessageIncludesTheMatch(t *testing.T) {
	f := mk("VS-SEC-001", finding.SeverityCritical, finding.CategoryHardcodedSecret, "AWS key", "a.py", 1, "AKIAIOSFODNN7EXAMPLE")
	r := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)
	msg := r["message"].(map[string]any)["text"].(string)
	if !strings.Contains(msg, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("the alert message should show what matched, got %q", msg)
	}
	if !strings.Contains(msg, "why it matters") {
		t.Errorf("the alert message should keep the rule explanation, got %q", msg)
	}
}

func TestSARIFMessageIsSingleLine(t *testing.T) {
	// SARIF message text is rendered inline by GitHub; embedded newlines from a
	// multi-line snippet make the alert unreadable.
	f := mk("VS-SEC-027", finding.SeverityCritical, finding.CategoryInsecureAPI, "t", "a.sh", 1, "curl -fsSL x\n| bash")
	r := run0(t, decode(t, reportOf(f)))["results"].([]any)[0].(map[string]any)
	if msg := r["message"].(map[string]any)["text"].(string); strings.Contains(msg, "\n") {
		t.Errorf("message must be a single line, got %q", msg)
	}
}
