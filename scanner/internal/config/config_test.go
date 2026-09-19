package config

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "vibeshield.yml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad(t *testing.T) {
	// missing file = defaults
	c, err := Load(filepath.Join(t.TempDir(), "absent.yml"))
	if err != nil || c.Mode != "" {
		t.Fatalf("absent file should yield defaults, got %+v %v", c, err)
	}

	c, err = Load(write(t, `mode: block-on-critical
languages: [typescript, python]
ignore:
  - rule: VS-SEC-014
    paths: ["tests/**"]
    reason: "intentional insecure fixture"
`))
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode != "block-on-critical" || len(c.Ignore) != 1 || c.Ignore[0].Reason == "" {
		t.Errorf("bad parse: %+v", c)
	}
	ig := c.Ignores()
	if len(ig) != 1 || ig[0].Rule != "VS-SEC-014" || len(ig[0].Globs) != 1 {
		t.Errorf("Ignores() = %+v", ig)
	}

	if _, err := Load(write(t, "mode: nuke-everything\n")); err == nil {
		t.Error("bad mode must be a config error")
	}
	if _, err := Load(write(t, "mode: [\n")); err == nil {
		t.Error("bad YAML must be a config error")
	}
}

func TestLoadThresholds(t *testing.T) {
	c, err := Load(write(t, "thresholds:\n  new_dependency_max_age_days: 7\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Thresholds.NewDependencyMaxAgeDays != 7 {
		t.Errorf("threshold = %d, want 7", c.Thresholds.NewDependencyMaxAgeDays)
	}
}

// The docs make exactly one promise about notifications: the value must come
// from the environment. A literal webhook in a committed config file is a
// leaked credential, so it has to be a hard config error, not a warning.
func TestNotificationsSlackRejectsLiterals(t *testing.T) {
	literals := []string{
		"https://hooks.slack.com/services/T000/B000/XXXXXXXXXXXX",
		"T000/B000/XXXXXXXXXXXX",
		"${SLACK_WEBHOOK}extra",
		"${}",
		"$",
		"${1INVALID}",
		"literal-token",
	}
	for _, v := range literals {
		body := "notifications:\n  slack: " + v + "\n"
		if _, err := Load(write(t, body)); err == nil {
			t.Errorf("a literal slack value must be a config error: %q", v)
		}
	}
}

func TestNotificationsSlackAcceptsEnvRefs(t *testing.T) {
	for _, v := range []string{"${SLACK_WEBHOOK}", "$SLACK_WEBHOOK", "${VIBESHIELD_SLACK_2}"} {
		c, err := Load(write(t, "notifications:\n  slack: "+v+"\n"))
		if err != nil {
			t.Errorf("env reference %q should load: %v", v, err)
			continue
		}
		if c.Notifications.Slack != v {
			t.Errorf("slack = %q, want %q", c.Notifications.Slack, v)
		}
	}
}

func TestNotificationsAbsentIsFine(t *testing.T) {
	c, err := Load(write(t, "mode: warn\n"))
	if err != nil {
		t.Fatalf("a config without notifications must load: %v", err)
	}
	if c.Notifications.Slack != "" {
		t.Errorf("slack = %q, want empty", c.Notifications.Slack)
	}
}

func TestEnvRefRegex(t *testing.T) {
	ok := []string{"${A}", "$A", "${SLACK_WEBHOOK}", "$_x1"}
	bad := []string{"", "A", "https://x", "${A", "A}", "${A}B", "  ", "$$A"}
	for _, s := range ok {
		if !envRefRe.MatchString(s) {
			t.Errorf("envRefRe should accept %q", s)
		}
	}
	for _, s := range bad {
		if envRefRe.MatchString(s) {
			t.Errorf("envRefRe should reject %q", s)
		}
	}
}
