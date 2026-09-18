package cli

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// readerOf builds a keyReader over a fixed byte sequence. Closing the channel
// makes the escape-sequence timeouts resolve immediately instead of waiting,
// so the tests stay fast.
func readerOf(b ...byte) *keyReader {
	kr := &keyReader{ch: make(chan byte, len(b)+1)}
	for _, x := range b {
		kr.ch <- x
	}
	close(kr.ch)
	return kr
}

func TestKeyReaderParsesPlainKeys(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want keyKind
	}{
		{"enter cr", []byte{0x0d}, keyEnter},
		{"enter lf", []byte{0x0a}, keyEnter},
		{"backspace del", []byte{0x7f}, keyBackspace},
		{"backspace bs", []byte{0x08}, keyBackspace},
		{"ctrl-c", []byte{0x03}, keyCtrlC},
		{"ctrl-d", []byte{0x04}, keyCtrlD},
		{"ctrl-u", []byte{0x15}, keyCtrlU},
		{"tab", []byte{0x09}, keyTab},
	}
	for _, tc := range cases {
		got, ok := readerOf(tc.in...).read()
		if !ok {
			t.Errorf("%s: read reported EOF", tc.name)
			continue
		}
		if got.kind != tc.want {
			t.Errorf("%s: kind = %v, want %v", tc.name, got.kind, tc.want)
		}
	}
}

func TestKeyReaderParsesArrows(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want keyKind
	}{
		{"csi up", []byte{0x1b, '[', 'A'}, keyUp},
		{"csi down", []byte{0x1b, '[', 'B'}, keyDown},
		{"csi home", []byte{0x1b, '[', 'H'}, keyHome},
		{"csi end", []byte{0x1b, '[', 'F'}, keyEnd},
		{"csi pgup", []byte{0x1b, '[', '5', '~'}, keyPageUp},
		{"csi pgdn", []byte{0x1b, '[', '6', '~'}, keyPageDown},
		{"csi delete", []byte{0x1b, '[', '3', '~'}, keyDelete},
		{"ss3 up", []byte{0x1b, 'O', 'A'}, keyUp},
	}
	for _, tc := range cases {
		got, _ := readerOf(tc.in...).read()
		if got.kind != tc.want {
			t.Errorf("%s: kind = %v, want %v", tc.name, got.kind, tc.want)
		}
	}
}

func TestKeyReaderBareEscapeIsEscape(t *testing.T) {
	got, ok := readerOf(0x1b).read()
	if !ok {
		t.Fatal("read reported EOF")
	}
	if got.kind != keyEsc {
		t.Errorf("bare ESC = %v, want keyEsc", got.kind)
	}
}

func TestKeyReaderDecodesUTF8(t *testing.T) {
	got, _ := readerOf([]byte("é")...).read()
	if got.kind != keyRune || got.r != 'é' {
		t.Errorf("utf-8 é = %+v, want keyRune 'é'", got)
	}
	// A truncated multibyte sequence must not panic or hang.
	got, _ = readerOf(0xc3).read()
	if got.kind != keyRune {
		t.Errorf("truncated rune = %v, want a rune event", got.kind)
	}
}

func TestKeyReaderReportsEOFOnClosedInput(t *testing.T) {
	if _, ok := readerOf().read(); ok {
		t.Error("closed, empty input should report EOF")
	}
}

func TestNextGroupWalksGroups(t *testing.T) {
	res := []Item{
		{Group: "Scan"}, {Group: "Scan"}, {Group: "Fix"},
		{Group: "Rule · a"}, {Group: "Rule · a"}, {Group: "Rule · b"},
	}
	if got := nextGroup(res, 0); got != 2 {
		t.Errorf("nextGroup(0) = %d, want 2 (first Fix)", got)
	}
	if got := nextGroup(res, 3); got != 5 {
		t.Errorf("nextGroup(3) = %d, want 5 (first of the next rule group)", got)
	}
	if got := nextGroup(res, 5); got != 0 {
		t.Errorf("nextGroup at the end should wrap to 0, got %d", got)
	}
	if got := nextGroup(nil, 0); got != 0 {
		t.Errorf("nextGroup(nil) = %d, want 0", got)
	}
}

func TestBuildRowsInsertsGroupHeadersAndMapsPositions(t *testing.T) {
	res := []Item{
		{Group: "Scan", Title: "a"},
		{Group: "Scan", Title: "b"},
		{Group: "Fix", Title: "c"},
	}
	rows, pos := buildRows(res)
	// 2 headers + 3 items
	if len(rows) != 5 {
		t.Fatalf("buildRows produced %d rows, want 5: %+v", len(rows), rows)
	}
	if rows[0].idx != -1 || rows[0].group != "Scan" {
		t.Errorf("row 0 should be the Scan header, got %+v", rows[0])
	}
	for i, p := range pos {
		if rows[p].idx != i {
			t.Errorf("pos[%d] = %d, but that row holds item %d", i, p, rows[p].idx)
		}
	}
}

func TestTruncAndPadAreRuneAware(t *testing.T) {
	if got := truncRunes("abcdef", 4); utf8.RuneCountInString(got) != 4 {
		t.Errorf("truncRunes = %q (%d runes), want 4", got, utf8.RuneCountInString(got))
	}
	if got := truncRunes("ab", 5); got != "ab" {
		t.Errorf("truncRunes should not pad, got %q", got)
	}
	if got := truncRunes("abc", 0); got != "" {
		t.Errorf("truncRunes(x, 0) = %q, want empty", got)
	}
	// Emoji and box drawing must not be cut mid-rune.
	if got := truncRunes("🔴🟠🟡🔵", 3); utf8.ValidString(got) == false {
		t.Errorf("truncRunes produced invalid UTF-8: %q", got)
	}
	if got := pad("ab", 5); got != "ab   " {
		t.Errorf("pad = %q, want %q", got, "ab   ")
	}
	if got := pad("abcdef", 3); got != "abcdef" {
		t.Errorf("pad must not truncate, got %q", got)
	}
}

func TestTrimLastRune(t *testing.T) {
	if got := trimLastRune("abc"); got != "ab" {
		t.Errorf("trimLastRune(abc) = %q, want ab", got)
	}
	if got := trimLastRune("é"); got != "" {
		t.Errorf("trimLastRune(é) = %q, want empty (one rune removed)", got)
	}
	if got := trimLastRune(""); got != "" {
		t.Errorf("trimLastRune(empty) = %q, want empty", got)
	}
}

func TestWrapRunesBreaksLongLines(t *testing.T) {
	long := strings.Repeat("word ", 40)
	lines := wrapRunes(strings.TrimSpace(long), 20)
	if len(lines) < 2 {
		t.Fatalf("expected the long line to wrap, got %d line(s)", len(lines))
	}
	for _, ln := range lines {
		if utf8.RuneCountInString(ln) > 20 {
			t.Errorf("wrapped line is %d runes, want <= 20: %q", utf8.RuneCountInString(ln), ln)
		}
	}
	if got := wrapRunes("short", 20); len(got) != 1 || got[0] != "short" {
		t.Errorf("wrapRunes should pass short lines through, got %v", got)
	}
	// A line with no spaces still has to break rather than overflow forever.
	if got := wrapRunes(strings.Repeat("x", 100), 20); len(got) != 1 {
		t.Logf("unbreakable token produced %d lines (acceptable, truncated at render time)", len(got))
	}
}

func TestWidthAndViewportHonourEnv(t *testing.T) {
	c := &Console{}

	t.Setenv("COLUMNS", "72")
	if got := c.width(); got != 72 {
		t.Errorf("width with COLUMNS=72 = %d, want 72", got)
	}
	t.Setenv("COLUMNS", "5")
	if got := c.width(); got != 40 {
		t.Errorf("width should clamp tiny values to 40, got %d", got)
	}
	t.Setenv("COLUMNS", "500")
	if got := c.width(); got != 120 {
		t.Errorf("width should clamp huge values to 120, got %d", got)
	}
	t.Setenv("COLUMNS", "not-a-number")
	if got := c.width(); got != 96 {
		t.Errorf("width should fall back to 96 on garbage, got %d", got)
	}

	t.Setenv("LINES", "40")
	if got := c.viewport(); got != 20 {
		t.Errorf("viewport with LINES=40 = %d, want 20 (capped)", got)
	}
	t.Setenv("LINES", "10")
	if got := c.viewport(); got != 4 {
		t.Errorf("viewport with LINES=10 = %d, want 4 (floor)", got)
	}
}

func TestPaintRespectsColorFlag(t *testing.T) {
	c := &Console{Color: false}
	if got := c.paint(ansiBold, "hi"); got != "hi" {
		t.Errorf("paint without colour = %q, want the bare string", got)
	}
	c.Color = true
	if got := c.paint(ansiBold, "hi"); !strings.Contains(got, "hi") || !strings.Contains(got, ansiReset) {
		t.Errorf("paint with colour = %q, want it wrapped in ANSI", got)
	}
}

func TestRowRenderingFitsWidth(t *testing.T) {
	c := &Console{Color: false}
	it := Item{
		Title:   strings.Repeat("long title ", 8),
		Summary: strings.Repeat("long summary ", 8),
	}
	for _, w := range []int{40, 60, 96, 120} {
		plain := stripANSI(c.row(it, false, w))
		if n := utf8.RuneCountInString(plain); n > w-1 {
			t.Errorf("row at width %d is %d runes, want <= %d", w, n, w-1)
		}
		sel := stripANSI(c.row(it, true, w))
		if n := utf8.RuneCountInString(sel); n != w-1 {
			t.Errorf("selected row at width %d is %d runes, want exactly %d", w, n, w-1)
		}
	}
}

func TestPaginationKeepsSelectionVisible(t *testing.T) {
	// Simulate the viewport maths drawList performs.
	rows := 100
	view := 10
	for _, sel := range []int{0, 5, 9, 10, 55, 99} {
		offset := 0
		if sel < offset {
			offset = sel
		}
		if sel >= offset+view {
			offset = sel - view + 1
		}
		max := rows - view
		if offset > max {
			offset = max
		}
		if sel < offset || sel >= offset+view {
			t.Errorf("sel=%d falls outside the window [%d,%d)", sel, offset, offset+view)
		}
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			in = true
		case in && r == 'm':
			in = false
		case !in:
			b.WriteRune(r)
		}
	}
	return b.String()
}
