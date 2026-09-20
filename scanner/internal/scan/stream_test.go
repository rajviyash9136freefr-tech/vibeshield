package scan

import (
	"context"
	"testing"
	"time"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

func TestStreamScan(t *testing.T) {
	pack, err := rules.LoadCore()
	if err != nil {
		t.Fatalf("LoadCore: %v", err)
	}

	testDir := writeTree(t, map[string]string{
		"server.py": "app.run(port=5057, debug=True)\n",
		"clean.ts":  "export const add = (a: number, b: number) => a + b;\n",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := Stream(ctx, testDir, pack, "test-version", Options{})

	var fileCount int
	var findingsCount int
	var completed bool

	for ev := range ch {
		switch ev.Kind {
		case EventFileScanned:
			fileCount++
		case EventFindingDiscovered:
			findingsCount++
			if ev.Finding == nil {
				t.Errorf("expected finding payload, got nil")
			}
		case EventScanComplete:
			completed = true
			if ev.Report == nil {
				t.Errorf("expected report on complete, got nil")
			}
		case EventScanFailed:
			t.Fatalf("unexpected scan failure: %v", ev.Error)
		}
	}

	if !completed {
		t.Errorf("expected EventScanComplete, but stream ended without it")
	}
	if findingsCount == 0 {
		t.Errorf("expected findings in testDir, got 0")
	}
}
