// Package finding defines the VibeShield Finding type (contracts/finding/
// schema.json, schema_version 1) plus snippet redaction and dismiss-hash
// helpers. The scanner is the emitter and therefore owns secret redaction —
// rule authors never redact (contracts/rulepack.md, engine semantics).
package finding

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// Categories from contracts/finding/schema.json.
const (
	CategoryHallucinatedPackage = "hallucinated-package"
	CategoryHardcodedSecret     = "hardcoded-secret"
	CategoryInsecureAPI         = "insecure-api"
	CategoryLicenseMissing      = "license-missing"
	CategoryInsecureDefault     = "insecure-default"
	CategoryDependencyRisk      = "dependency-risk"
	CategoryPromptInjection     = "prompt-injection"
)

// Severity tokens (lowercase wire format; UI displays uppercase).
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// AI-origin values.
const (
	OriginConfirmed = "confirmed"
	OriginLikely    = "likely"
	OriginUnknown   = "unknown"
)

// Bullet used for redaction (U+2022) per the finding schema.
const RedactionChar = '•'

// Finding is one contracts/finding/schema.json object. additionalProperties
// is false in the schema, so this struct must carry exactly the schema keys.
type Finding struct {
	SchemaVersion int            `json:"schema_version"`
	RuleID        string         `json:"rule_id"`
	Severity      string         `json:"severity"`
	Category      string         `json:"category"`
	Title         string         `json:"title"`
	Message       string         `json:"message"`
	File          string         `json:"file"`
	Line          int            `json:"line,omitempty"`
	EndLine       int            `json:"end_line,omitempty"`
	Column        int            `json:"column,omitempty"`
	Snippet       string         `json:"snippet,omitempty"`
	Fix           string         `json:"fix,omitempty"`
	AIOrigin      string         `json:"ai_origin"`
	Confidence    float64        `json:"confidence,omitempty"`
	Dismissable   *bool          `json:"dismissable,omitempty"`
	DismissHash   string         `json:"dismiss_hash,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// New returns a Finding with the mandatory envelope fields filled in.
func New(ruleID, severity, category, title, message, file string) *Finding {
	t := true
	return &Finding{
		SchemaVersion: 1,
		RuleID:        ruleID,
		Severity:      severity,
		Category:      category,
		Title:         title,
		Message:       message,
		File:          relPath(file),
		AIOrigin:      OriginUnknown,
		Dismissable:   &t,
	}
}

// Finalize computes the dismiss hash. Call after line/snippet are set.
func (f *Finding) Finalize() {
	f.DismissHash = DismissHash(f.RuleID, f.File, f.Line, f.Snippet)
}

// relPath normalizes to repo-relative POSIX (schema: "Repo-relative POSIX path").
func relPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	return p
}

// FinalizePath normalizes File to POSIX repo-relative form.
func (f *Finding) FinalizePath() { f.File = relPath(f.File) }

var wsRe = regexp.MustCompile(`\s+`)

// NormalizeSnippet lowercases, trims and collapses whitespace so that
// dismiss hashes survive formatting-only edits.
func NormalizeSnippet(s string) string {
	return strings.ToLower(wsRe.ReplaceAllString(strings.TrimSpace(s), " "))
}

// DismissHash is a stable short hash of rule_id + file + line + normalized
// snippet, used by /vibeshield accept commands and for dedupe.
func DismissHash(ruleID, file string, line int, snippet string) string {
	h := sha256.New()
	h.Write([]byte(ruleID))
	h.Write([]byte{'\n'})
	h.Write([]byte(relPath(file)))
	h.Write([]byte{'\n'})
	h.Write([]byte{byte(line), byte(line >> 8), byte(line >> 16), byte(line >> 24)})
	h.Write([]byte{'\n'})
	h.Write([]byte(NormalizeSnippet(snippet)))
	return hex.EncodeToString(h.Sum(nil)[:6]) // 12 hex chars — "short hash"
}

// RedactRegion masks the matched secret region inside line: keeps a 4-char
// prefix and 4-char suffix of the secret, masks the middle with U+2022.
// Regions of 8 chars or fewer are not maskable (nothing to hide) so they are
// left intact. start/end are byte offsets into line.
func RedactRegion(line string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(line) || end < 0 {
		end = len(line)
	}
	if end-start <= 8 {
		return line
	}
	secret := line[start:end]
	var b strings.Builder
	b.WriteString(line[:start])
	b.WriteString(secret[:4])
	b.WriteString(strings.Repeat(string(RedactionChar), runeCount(secret[4:len(secret)-4])))
	b.WriteString(secret[len(secret)-4:])
	b.WriteString(line[end:])
	return b.String()
}

func runeCount(s string) int {
	// mask per byte of the middle so the width of the original is roughly
	// preserved for ASCII (the common case for API keys).
	return len(s)
}

// SecretType infers a stable secret type label for metadata from a known
// prefix; falls back to "api_key".
func SecretType(secret string) string {
	switch {
	case strings.HasPrefix(secret, "sk-ant-"):
		return "anthropic_api_key"
	case strings.HasPrefix(secret, "sk-proj-"):
		return "openai_project_api_key"
	case strings.HasPrefix(secret, "sk-"):
		return "openai_api_key"
	case strings.HasPrefix(secret, "ghp_"), strings.HasPrefix(secret, "gho_"),
		strings.HasPrefix(secret, "ghu_"), strings.HasPrefix(secret, "ghs_"),
		strings.HasPrefix(secret, "ghr_"):
		return "github_token"
	case strings.HasPrefix(secret, "AKIA"):
		return "aws_access_key_id"
	case strings.HasPrefix(secret, "xoxb-"), strings.HasPrefix(secret, "xoxa-"),
		strings.HasPrefix(secret, "xoxp-"), strings.HasPrefix(secret, "xoxr-"),
		strings.HasPrefix(secret, "xoxs-"):
		return "slack_token"
	default:
		return "api_key"
	}
}
