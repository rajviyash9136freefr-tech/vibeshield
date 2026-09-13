// Package config loads vibeshield.yml (contracts/cli.md schema). Unknown
// fields are ignored on purpose for forward-compat, but malformed YAML or
// bad mode values are config errors (exit 2).
package config

import (
	"fmt"
	"os"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
	"gopkg.in/yaml.v3"
)

// Config is the vibeshield.yml document (subset the minimal engine honours).
type Config struct {
	Mode       string   `yaml:"mode"`
	Languages  []string `yaml:"languages"`
	Ignore     []ignore `yaml:"ignore"`
	Thresholds struct {
		NewDependencyMaxAgeDays int `yaml:"new_dependency_max_age_days"`
	} `yaml:"thresholds"`
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
	return &c, nil
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
