package scandiff

import (
	"testing"
)

const sample = `diff --git a/src/app.ts b/src/app.ts
index 1111111..2222222 100644
--- a/src/app.ts
+++ b/src/app.ts
@@ -0,0 +1,2 @@
+const key = "AKIAFAKEFAKEFAKEFAKE";
+export default key;
@@ -10,0 +12,1 @@
+added line
`

const deleted = `diff --git a/x.js b/x.js
--- a/x.js
+++ /dev/null
@@ -1,2 +0,0 @@
-const gone = 1;
-const gone2 = 2;
`

func TestParseAddedLines(t *testing.T) {
	d := newDiff("origin/main")
	d.parse(sample)
	if len(d.Files) != 1 || d.Files[0] != "src/app.ts" {
		t.Fatalf("files = %v, want [src/app.ts]", d.Files)
	}
	for _, ln := range []int{1, 2, 12} {
		if !d.Has("src/app.ts", ln) {
			t.Errorf("expected added line %d", ln)
		}
	}
	if d.Has("src/app.ts", 3) {
		t.Error("line 3 is not in the diff")
	}
}

func TestDeletedFileProducesNoLines(t *testing.T) {
	d := newDiff("HEAD~1")
	d.parse(deleted)
	if len(d.Files) != 0 {
		t.Errorf("deletions must not add files, got %v", d.Files)
	}
}

func TestFromStdinText(t *testing.T) {
	d := FromStdinText(sample)
	if d.Ref != "stdin" || !d.Has("src/app.ts", 1) {
		t.Errorf("stdin parse broken: %+v", d)
	}
}
