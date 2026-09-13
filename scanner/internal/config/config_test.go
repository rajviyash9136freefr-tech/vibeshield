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
