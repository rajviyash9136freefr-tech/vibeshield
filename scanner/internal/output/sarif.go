// SARIF 2.1.0 output (contracts/cli.md: "--format sarif targets GitHub
// code-scanning uploads"). The document is what github/codeql-action/
// upload-sarif consumes, so findings land in a repository's Security tab as
// code-scanning alerts with the rule text, the one-line fix and a stable
// fingerprint that survives reformatting.
package output

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
)

// SARIFVersion is the schema revision we emit. GitHub code scanning requires
// 2.1.0; there is no 2.2.
const SARIFVersion = "2.1.0"

// sarifSchema is the canonical schema URI for the version above.
const sarifSchema = "https://json.schemastore.org/sarif-2.1.0.json"

// sarifInformationURI is the tool's home, shown on every alert.
const sarifInformationURI = "https://github.com/rajviyash9136freefr-tech/vibeshield"

// sarifRulesURI documents the rule packs the ids refer to.
const sarifRulesURI = sarifInformationURI + "/blob/main/rules/core"

// fingerprintKey names the partial fingerprint. SARIF allows any key; GitHub
// uses the value to decide whether an alert is the same one it saw before, so
// a stable key means fixing one line does not re-open every other alert.
const fingerprintKey = "vibeshieldFindingHash/v1"

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool        sarifTool         `json:"tool"`
	Invocations []sarifInvocation `json:"invocations,omitempty"`
	Results     []sarifResult     `json:"results"`
}

type sarifInvocation struct {
	// A scan that completed always exits "successfully" in the SARIF sense:
	// findings are results, not tool failures. The gate exit code is a
	// separate concern (contracts/cli.md exit-code law).
	ExecutionSuccessful bool `json:"executionSuccessful"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	ShortDescription     sarifText      `json:"shortDescription"`
	FullDescription      sarifText      `json:"fullDescription"`
	Help                 sarifHelp      `json:"help"`
	HelpURI              string         `json:"helpUri"`
	DefaultConfiguration sarifConfig    `json:"defaultConfiguration"`
	Properties           map[string]any `json:"properties,omitempty"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifHelp struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

type sarifConfig struct {
	Level string `json:"level"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId"`
	RuleIndex           int               `json:"ruleIndex"`
	Level               string            `json:"level"`
	Message             sarifText         `json:"message"`
	Locations           []sarifLocation   `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
	Properties          map[string]any    `json:"properties,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           *sarifRegion  `json:"region,omitempty"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int        `json:"startLine,omitempty"`
	StartColumn int        `json:"startColumn,omitempty"`
	EndLine     int        `json:"endLine,omitempty"`
	Snippet     *sarifText `json:"snippet,omitempty"`
}

// sarifLevel maps a severity token to a SARIF result level. GitHub renders
// these as its own error/warning/note severities in the Security tab, so the
// mapping is what decides how loud an alert looks.
func sarifLevel(sev string) string {
	switch sev {
	case finding.SeverityCritical, finding.SeverityHigh:
		return "error"
	case finding.SeverityMedium:
		return "warning"
	default: // low, info
		return "note"
	}
}

// SARIF writes the contracts/cli.md sarif document. Rules are emitted once
// each, sorted by id, and every result carries a ruleIndex into that array —
// GitHub rejects results whose ruleId is absent from the driver.
func SARIF(w io.Writer, rep *scan.Report) error {
	fs := append([]*finding.Finding{}, rep.Findings...)
	sortFindings(fs)

	rules := collectRules(fs)
	index := make(map[string]int, len(rules))
	for i, r := range rules {
		index[r.ID] = i
	}

	results := make([]sarifResult, 0, len(fs))
	for _, f := range fs {
		results = append(results, sarifResult{
			RuleID:              f.RuleID,
			RuleIndex:           index[f.RuleID],
			Level:               sarifLevel(f.Severity),
			Message:             sarifText{Text: resultMessage(f)},
			Locations:           []sarifLocation{{PhysicalLocation: physicalLocation(f)}},
			PartialFingerprints: map[string]string{fingerprintKey: f.DismissHash},
			Properties:          resultProperties(f),
		})
	}

	doc := sarifLog{
		Schema:  sarifSchema,
		Version: SARIFVersion,
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name:           "VibeShield",
				Version:        rep.Version,
				InformationURI: sarifInformationURI,
				Rules:          rules,
			}},
			Invocations: []sarifInvocation{{ExecutionSuccessful: true}},
			Results:     results,
		}},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// collectRules builds one reporting descriptor per distinct rule that fired,
// sorted by id so the output is byte-stable between runs.
func collectRules(fs []*finding.Finding) []sarifRule {
	seen := map[string]sarifRule{}
	for _, f := range fs {
		if _, ok := seen[f.RuleID]; ok {
			continue
		}
		seen[f.RuleID] = sarifRule{
			ID:                   f.RuleID,
			Name:                 f.RuleID, // SARIF names must not contain whitespace
			ShortDescription:     sarifText{Text: f.Title},
			FullDescription:      sarifText{Text: f.Message},
			Help:                 sarifHelp{Text: helpText(f), Markdown: helpMarkdown(f)},
			HelpURI:              sarifRulesURI,
			DefaultConfiguration: sarifConfig{Level: sarifLevel(f.Severity)},
			Properties: map[string]any{
				"category": f.Category,
				"severity": f.Severity,
				"tags":     []string{"security", "ai-generated-code", f.Category},
			},
		}
	}
	out := make([]sarifRule, 0, len(seen))
	for _, r := range seen {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// resultMessage is the alert text: the rule's explanation plus what was
// actually matched, so the Security tab reads without opening the diff.
func resultMessage(f *finding.Finding) string {
	msg := f.Message
	if f.Snippet != "" {
		msg = strings.TrimSpace(msg) + " Matched: " + oneLine(f.Snippet, 160)
	}
	return msg
}

func helpText(f *finding.Finding) string {
	var b strings.Builder
	b.WriteString(f.Message)
	if f.Fix != "" {
		b.WriteString("\n\nFix: ")
		b.WriteString(f.Fix)
	}
	return b.String()
}

func helpMarkdown(f *finding.Finding) string {
	var b strings.Builder
	b.WriteString(f.Message)
	if f.Fix != "" {
		b.WriteString("\n\n**Fix:** ")
		b.WriteString(f.Fix)
	}
	if f.Category != "" {
		b.WriteString("\n\nCategory: `")
		b.WriteString(f.Category)
		b.WriteString("`")
	}
	return b.String()
}

// sarifURI guarantees a repository-relative POSIX uri. SARIF consumers match
// artifactLocation.uri against the checked-out tree, so an absolute or
// backslash path produces an alert with no file attached — and no error. The
// scanner already emits relative paths; doing it here makes the guarantee
// local to the emitter instead of an invariant it merely hopes for.
func sarifURI(p string) string {
	p = strings.ReplaceAll(p, `\`, "/")
	if len(p) >= 2 && p[1] == ':' && isASCIILetter(p[0]) {
		p = p[2:] // strip a Windows drive prefix: C:/dev/x -> /dev/x
	}
	p = strings.TrimLeft(p, "/")
	p = strings.TrimPrefix(p, "./")
	if p == "" {
		return "."
	}
	return p
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// physicalLocation maps the finding's repo-relative path and line into SARIF.
// A file-level finding (Line 0) has no region — emitting startLine 0 would be
// invalid SARIF, so the region is omitted instead.
func physicalLocation(f *finding.Finding) sarifPhysical {
	loc := sarifPhysical{ArtifactLocation: sarifArtifact{URI: sarifURI(f.File)}}
	if f.Line <= 0 {
		return loc
	}
	region := &sarifRegion{StartLine: f.Line, StartColumn: f.Column}
	if region.StartColumn <= 0 {
		region.StartColumn = 1
	}
	if f.EndLine > f.Line {
		region.EndLine = f.EndLine
	}
	if f.Snippet != "" {
		region.Snippet = &sarifText{Text: oneLine(f.Snippet, 400)}
	}
	loc.Region = region
	return loc
}

// resultProperties carries the contract fields SARIF has no slot for, so an
// alert in the Security tab still shows the AI-origin tag and the fix.
func resultProperties(f *finding.Finding) map[string]any {
	props := map[string]any{
		"severity": f.Severity,
		"category": f.Category,
		"aiOrigin": f.AIOrigin,
	}
	if f.Fix != "" {
		props["fix"] = f.Fix
	}
	if f.Confidence > 0 {
		props["confidence"] = f.Confidence
	}
	if f.DismissHash != "" {
		props["dismissHash"] = f.DismissHash
	}
	for k, v := range f.Metadata {
		// Namespace contract metadata so it cannot collide with our keys.
		props["vibeshield."+k] = v
	}
	return props
}
