package tui

import (
	"strings"
	"testing"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

func TestAppModelViews(t *testing.T) {
	pack, err := rules.LoadCore()
	if err != nil {
		t.Fatalf("LoadCore: %v", err)
	}

	app := New(".", pack, "3.1.0-test")

	// 1. Initial view is Home
	if app.view != ViewHome {
		t.Errorf("expected ViewHome, got %v", app.view)
	}
	homeView := app.View()
	if !strings.Contains(homeView, "VIBESHIELD") {
		t.Errorf("expected home view to contain VIBESHIELD banner, got: %s", homeView)
	}

	// 2. Add sample findings and test Findings view
	f := finding.New("VS-SEC-014", "high", "insecure-default", "Flask debug=True", "debug exposes RCE", "server.py")
	f.Line = 42
	f.Snippet = "app.run(debug=True)"
	f.Fix = "app.run(debug=False)"
	f.Finalize()

	app.findings = []*finding.Finding{f}
	app.view = ViewFindings
	findingsView := app.View()
	if !strings.Contains(findingsView, "AUDIT RESULTS") || !strings.Contains(findingsView, "VS-SEC-014") {
		t.Errorf("expected findings view to render findings, got: %s", findingsView)
	}

	// 3. Test Detail view
	app.view = ViewDetail
	detailView := app.View()
	if !strings.Contains(detailView, "FINDING DETAIL") || !strings.Contains(detailView, "server.py:42") {
		t.Errorf("expected detail view to render details, got: %s", detailView)
	}

	// 4. Test Help view
	app.view = ViewHelp
	helpView := app.View()
	if !strings.Contains(helpView, "COMMAND GUIDE") {
		t.Errorf("expected help view to render command guide, got: %s", helpView)
	}
}
