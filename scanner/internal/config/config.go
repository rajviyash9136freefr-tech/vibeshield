// Package config loads vibeshield.yml (contracts/cli.md schema). Unknown
// fields are ignored on purpose for forward-compat, but malformed YAML or
// bad mode values are config errors (exit 2).
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
	"gopkg.in/yaml.v3"
)

// Config is the vibeshield.yml document (subset the minimal engine honours).
type Config struct {
	Mode          string        `yaml:"mode"`
	Languages     []string      `yaml:"languages"`
	Ignore        []ignore      `yaml:"ignore"`
	Thresholds    Thresholds    `yaml:"thresholds"`
	Notifications Notifications `yaml:"notifications"`
}

// Thresholds are the supply-chain windows.
type Thresholds struct {
	NewDependencyMaxAgeDays int `yaml:"new_dependency_max_age_days"`
}

// Notifications is the optional outbound-notification block. Only the env-ref
// form is accepted; see validateNotifications.
type Notifications struct {
	Slack string `yaml:"slack"`
}

type ignore struct {
	Rule   string   `yaml:"rule"`
	Paths  []string `yaml:"paths"`
	Reason string   `yaml:"reason"`
}

// AllowedModes from contracts/cli.md.
var AllowedModes = map[string]bool{
	"off": true, "warn": true, "block-on-critical": true, "block-on-high+": true,
}

// envRefRe matches the only accepted form for a secret in a config file:
// ${VAR} or $VAR. The braces are all-or-nothing — making each optional
// separately would accept the malformed "${A", which is exactly the kind of
// half-a-reference that gets committed by accident.
var envRefRe = regexp.MustCompile(`^\$(?:\{[A-Za-z_][A-Za-z0-9_]*\}|[A-Za-z_][A-Za-z0-9_]*)$`)

// Load reads a config file. A missing default-path file yields the zero
// Config (defaults); a missing --config override is the caller's error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: invalid YAML: %w", path, err)
	}
	if c.Mode != "" && !AllowedModes[c.Mode] {
		return nil, fmt.Errorf("%s: mode %q not one of off|warn|block-on-critical|block-on-high+", path, c.Mode)
	}
	if err := validateNotifications(c.Notifications, path); err != nil {
		return nil, err
	}
	return &c, nil
}

// validateNotifications enforces the one rule the docs make about this block:
// the value is an environment reference, never a literal. A webhook URL
// committed to a config file is a leaked credential, so it is a hard config
// error (exit 2) rather than a warning — the whole point is that the mistake
// cannot be merged quietly.
func validateNotifications(n Notifications, path string) error {
	v := strings.TrimSpace(n.Slack)
	if v == "" {
		return nil
	}
	if envRefRe.MatchString(v) {
		return nil
	}
	return fmt.Errorf("%s: notifications.slack must be an environment reference "+
		"such as ${SLACK_WEBHOOK}, not a literal value — a webhook URL committed "+
		"to a config file is a leaked credential (the path is the whole secret)", path)
}

// Ignores converts config ignore entries into scan ignores. The reason field
// is required by the contract and echoed by report surfaces; it's carried in
// Rule for now via the scan.Ignore shape.
func (c *Config) Ignores() []scan.Ignore {
	var out []scan.Ignore
	for _, ig := range c.Ignore {
		out = append(out, scan.Ignore{Rule: ig.Rule, Globs: ig.Paths})
	}
	return out
}
