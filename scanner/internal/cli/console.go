package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// errNotTerminal is returned by makeRaw when stdin is a pipe or a file, so the
// caller can fall back to a plain listing instead of hanging on a read.
var errNotTerminal = errors.New("stdin is not a terminal")

// ANSI control sequences. Colour is dropped when the writer is not a TTY or
// NO_COLOR is set, but cursor control is always used (we only get here after
// makeRaw succeeded, which means we do have a real terminal).
const (
	ansiReset      = "\x1b[0m"
	ansiBold       = "\x1b[1m"
	ansiDim        = "\x1b[2m"
	ansiReverse    = "\x1b[7m"
	ansiCyan       = "\x1b[36m"
	ansiGreen      = "\x1b[32m"
	ansiYellow     = "\x1b[33m"
	ansiGray       = "\x1b[90m"
	ansiClearLine  = "\r\x1b[2K"
	ansiHideCursor = "\x1b[?25l"
	ansiShowCursor = "\x1b[?25h"
)

// Console is the v2 interactive menu: one search box over every action, rule
// and agent recipe the scanner knows about.
type Console struct {
	In      *os.File
	Out     io.Writer
	Items   []Item
	Version string
	// Exec executes a CLI action; it is wired to the real command dispatcher
	// so the menu can never drift from the documented flags.
	Exec  func(args []string) int
	Color bool
}

// state carries what has to survive a redraw: how many lines we painted last
// time (to move the cursor back up) and how to restore the terminal.
type state struct {
	lines   int
	restore func()
}

// Run drives the menu until the user quits or picks an action. It returns the
// exit code of whatever it ran (0 when the user simply quit).
func (c *Console) Run() int {
	restore, err := makeRaw(c.In)
	if err != nil {
		c.plain()
		return 0
	}
	st := &state{restore: restore}
	defer func() {
		c.clear(st)
		fmt.Fprint(c.Out, ansiShowCursor)
		if st.restore != nil {
			st.restore()
		}
	}()
	fmt.Fprint(c.Out, ansiHideCursor)

	kr := newKeyReader(c.In)
	query := ""
	sel, offset := 0, 0
	var detail *Item
	detailOff := 0

	redraw := func() []Item {
		res := Rank(query, c.Items)
		if detail != nil {
			c.drawDetail(*detail, detailOff, st)
			return res
		}
		if sel >= len(res) {
			sel = len(res) - 1
		}
		if sel < 0 {
			sel = 0
		}
		c.drawList(query, res, sel, offset, st)
		return res
	}

	res := redraw()
	for {
		ev, ok := kr.read()
		if !ok {
			return 0
		}

		// Detail pane swallows navigation keys and returns on Esc/Enter.
		if detail != nil {
			switch ev.kind {
			case keyCtrlC, keyCtrlD:
				return 0
			case keyUp:
				if detailOff > 0 {
					detailOff--
				}
			case keyDown:
				detailOff++
			case keyPageUp:
				detailOff -= c.viewport()
				if detailOff < 0 {
					detailOff = 0
				}
			case keyPageDown:
				detailOff += c.viewport()
			case keyEsc, keyEnter, keyBackspace, keyLeft:
				detail, detailOff = nil, 0
			}
			redraw()
			continue
		}

		switch ev.kind {
		case keyCtrlC, keyCtrlD:
			return 0

		case keyEsc:
			if query != "" {
				query, sel, offset = "", 0, 0
				redraw()
				continue
			}
			return 0

		case keyUp:
			if sel > 0 {
				sel--
			}
		case keyDown:
			if sel < len(res)-1 {
				sel++
			}
		case keyHome:
			sel = 0
		case keyEnd:
			if len(res) > 0 {
				sel = len(res) - 1
			}
		case keyPageUp:
			sel -= c.viewport()
			if sel < 0 {
				sel = 0
			}
		case keyPageDown:
			sel += c.viewport()
			if len(res) > 0 && sel > len(res)-1 {
				sel = len(res) - 1
			}
		case keyTab:
			sel = nextGroup(res, sel)

		case keyCtrlU:
			query, sel, offset = "", 0, 0

		case keyBackspace:
			if query != "" {
				query = trimLastRune(query)
				sel, offset = 0, 0
			}

		case keyRune:
			query += string(ev.r)
			sel, offset = 0, 0

		case keyEnter:
			if len(res) == 0 {
				continue
			}
			it := res[sel]
			if it.Kind == KindInfo {
				detail, detailOff = &it, 0
				redraw()
				continue
			}
			// Action: hand the terminal back, run it, and let the process end
			// with the command's own exit code. The console is a launcher, so
			// it never fights the command for stdin.
			c.clear(st)
			fmt.Fprint(c.Out, ansiShowCursor)
			if st.restore != nil {
				st.restore()
				st.restore = nil
			}
			fmt.Fprintf(c.Out, "\n  %s\n\n", c.paint(ansiCyan, "$ vibeshield "+strings.Join(it.Args, " ")))
			if c.Exec == nil {
				return 0
			}
			return c.Exec(it.Args)
		}
		res = redraw()
	}
}

// nextGroup moves the selection to the first entry of the following group, so
// Tab walks Actions → Rules → Agents without typing a query.
func nextGroup(res []Item, sel int) int {
	if len(res) == 0 {
		return 0
	}
	cur := res[sel].Group
	for i := sel + 1; i < len(res); i++ {
		if res[i].Group != cur {
			return i
		}
	}
	return 0
}

// --- drawing ---------------------------------------------------------------

type rowLine struct {
	group string // set for a group header line
	item  Item
	idx   int // index into the ranked slice, -1 for headers
}

func buildRows(res []Item) ([]rowLine, []int) {
	rows := make([]rowLine, 0, len(res)+8)
	pos := make([]int, len(res))
	last := ""
	for i, it := range res {
		if it.Group != last {
			rows = append(rows, rowLine{group: it.Group, idx: -1})
			last = it.Group
		}
		pos[i] = len(rows)
		rows = append(rows, rowLine{item: it, idx: i})
	}
	return rows, pos
}

func (c *Console) drawList(query string, res []Item, sel, offset int, st *state) {
	w := c.width()
	rows, pos := buildRows(res)
	selLine := 0
	if len(res) > 0 {
		selLine = pos[sel]
	}
	view := c.viewport()

	// Keep the selected row inside the window.
	if selLine < offset {
		offset = selLine
	}
	if selLine >= offset+view {
		offset = selLine - view + 1
	}
	max := len(rows) - view
	if max < 0 {
		max = 0
	}
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}

	out := make([]string, 0, view+6)
	out = append(out, c.header(w))
	out = append(out, c.searchLine(query, w))
	out = append(out, c.rule(w))

	if len(res) == 0 {
		out = append(out, c.paint(ansiYellow, "    no match — backspace to widen the search"))
	} else {
		for i := offset; i < len(rows) && i < offset+view; i++ {
			if rows[i].idx < 0 {
				out = append(out, c.paint(ansiCyan, "  "+strings.ToUpper(rows[i].group)))
				continue
			}
			out = append(out, c.row(rows[i].item, rows[i].idx == sel, w))
		}
	}

	out = append(out, c.rule(w))
	out = append(out, c.footer(len(res), sel, w))
	c.paintBlock(out, st)
}

func (c *Console) drawDetail(it Item, off int, st *state) {
	w := c.width()
	view := c.viewport() + 2

	body := strings.Split(strings.ReplaceAll(it.Body, "\r\n", "\n"), "\n")
	wrapped := make([]string, 0, len(body)+4)
	for _, ln := range body {
		if strings.TrimSpace(ln) == "" {
			wrapped = append(wrapped, "")
			continue
		}
		wrapped = append(wrapped, wrapRunes(ln, w-6)...)
	}
	maxOff := len(wrapped) - view
	if maxOff < 0 {
		maxOff = 0
	}
	if off > maxOff {
		off = maxOff
	}
	if off < 0 {
		off = 0
	}

	out := make([]string, 0, view+6)
	out = append(out, c.header(w))
	out = append(out, "  "+c.paint(ansiBold, it.Title))
	out = append(out, "  "+c.paint(ansiGray, it.Group))
	out = append(out, c.rule(w))
	for i := off; i < len(wrapped) && i < off+view; i++ {
		out = append(out, "  "+wrapped[i])
	}
	for len(out) < view+4 {
		out = append(out, "")
	}
	out = append(out, c.rule(w))
	out = append(out, c.footerDetail(off, maxOff, w))
	c.paintBlock(out, st)
}

func (c *Console) header(w int) string {
	title := c.paint(ansiBold, "VibeShield "+c.Version)
	sub := fmt.Sprintf("%d entries · fully offline", len(c.Items))
	gap := w - utf8.RuneCountInString("VibeShield "+c.Version) - len(sub) - 4
	if gap < 1 {
		gap = 1
	}
	return "  " + title + strings.Repeat(" ", gap) + c.paint(ansiGray, sub)
}

func (c *Console) searchLine(query string, w int) string {
	prompt := c.paint(ansiCyan, "  ❯ ")
	if query == "" {
		return prompt + c.paint(ansiGray, "type to search actions, rules and agent setup")
	}
	return prompt + c.paint(ansiBold, query) + c.paint(ansiGray, "▏")
}

func (c *Console) row(it Item, selected bool, w int) string {
	titleW := 44
	if w < 92 {
		titleW = 34
	}
	prefix := "    "
	if selected {
		prefix = "  ▸ "
	}
	title := pad(truncRunes(it.Title, titleW), titleW)
	summary := truncRunes(it.Summary, w-titleW-8)
	line := truncRunes(prefix+title+"  "+summary, w-1)
	if selected {
		return ansiReverse + pad(line, w-1) + ansiReset
	}
	return prefix + c.paint(ansiGray, truncRunes(title+"  "+summary, w-5))
}

func (c *Console) footer(matches, sel, w int) string {
	hint := "  ↑↓ move · ⏎ open · tab next group · esc clear · ctrl+c quit"
	count := fmt.Sprintf("%d match", matches)
	if matches != 1 {
		count += "es"
	}
	if matches > 0 {
		count = fmt.Sprintf("%d/%d · %s", sel+1, matches, count)
	}
	gap := w - utf8.RuneCountInString(hint) - len(count) - 4
	if gap < 1 {
		gap = 1
	}
	return c.paint(ansiGray, hint) + strings.Repeat(" ", gap) + c.paint(ansiGray, count)
}

func (c *Console) footerDetail(off, maxOff, w int) string {
	hint := "  ↑↓ scroll · ⏎/esc back to results · ctrl+c quit"
	pos := fmt.Sprintf("%d/%d", off+1, maxOff+1)
	gap := w - utf8.RuneCountInString(hint) - len(pos) - 4
	if gap < 1 {
		gap = 1
	}
	return c.paint(ansiGray, hint) + strings.Repeat(" ", gap) + c.paint(ansiGray, pos)
}

func (c *Console) rule(w int) string {
	if w > 4 {
		return c.paint(ansiGray, strings.Repeat("─", w-2))
	}
	return ""
}

// paintBlock redraws the frame in place: move up over the previous frame, then
// rewrite every line with the line cleared first. When the new frame is
// shorter than the old one the leftover rows are blanked rather than left
// behind, which is what stops the menu smearing on filter changes.
func (c *Console) paintBlock(lines []string, st *state) {
	total := len(lines)
	if st.lines > total {
		total = st.lines
	}
	var b strings.Builder
	if st.lines > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", st.lines)
	}
	for i := 0; i < total; i++ {
		b.WriteString(ansiClearLine)
		if i < len(lines) {
			b.WriteString(lines[i])
		}
		b.WriteString("\n")
	}
	st.lines = total
	fmt.Fprint(c.Out, b.String())
}

// clear erases the frame and parks the cursor back at its first row, so
// whatever the launched command prints next starts on a clean screen.
func (c *Console) clear(st *state) {
	if st.lines == 0 {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\x1b[%dA", st.lines)
	for i := 0; i < st.lines; i++ {
		b.WriteString(ansiClearLine)
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\x1b[%dA", st.lines)
	st.lines = 0
	fmt.Fprint(c.Out, b.String())
}

// --- non-interactive fallback ----------------------------------------------

// plain is what a pipe gets: the same catalog, one line each, no cursor games.
func (c *Console) plain() {
	fmt.Fprintf(c.Out, "vibeshield %s — %d searchable entries\n\n", c.Version, len(c.Items))
	last := ""
	for _, it := range c.Items {
		if it.Group != last {
			fmt.Fprintf(c.Out, "%s\n", strings.ToUpper(it.Group))
			last = it.Group
		}
		fmt.Fprintf(c.Out, "  %-46s %s\n", truncRunes(it.Title, 46), it.Summary)
	}
	fmt.Fprintln(c.Out, "\nRun `vibeshield search <query>` to search these, or `vibeshield help` for commands.")
}

// --- geometry ---------------------------------------------------------------

// width is the render width. COLUMNS is honoured whenever it parses, then the
// result is clamped: below 40 the frame stops being readable, and past 120 the
// eye has to travel too far to pair a title with its summary.
func (c *Console) width() int {
	w := 96
	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			w = n
		}
	}
	if w < 40 {
		w = 40
	}
	if w > 120 {
		w = 120
	}
	return w
}

// viewport is how many result rows fit. We do not query the terminal for its
// size (that needs another platform syscall); LINES is honoured when set and
// the default is comfortable on a standard 24-row terminal.
func (c *Console) viewport() int {
	h := 16
	if v := os.Getenv("LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 8 {
			h = n - 8
		}
	}
	if h < 4 {
		h = 4
	}
	if h > 20 {
		h = 20
	}
	return h
}

// --- text helpers -----------------------------------------------------------

func (c *Console) paint(code, s string) string {
	if !c.Color {
		return s
	}
	return code + s + ansiReset
}

func truncRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n-1]) + "…"
}

func pad(s string, n int) string {
	d := n - utf8.RuneCountInString(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

func trimLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

// wrapRunes breaks a line at width runes, preferring a space break.
func wrapRunes(s string, width int) []string {
	if width < 8 {
		width = 8
	}
	if utf8.RuneCountInString(s) <= width {
		return []string{s}
	}
	var out []string
	indent := ""
	if n := len(s) - len(strings.TrimLeft(s, " ")); n > 0 {
		indent = strings.Repeat(" ", n)
	}
	words := strings.Fields(s)
	cur := ""
	for _, word := range words {
		if cur == "" {
			cur = word
			continue
		}
		if utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(word) > width {
			out = append(out, cur)
			cur = indent + word
			continue
		}
		cur += " " + word
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// --- keys -------------------------------------------------------------------

type keyKind int

const (
	keyUnknown keyKind = iota
	keyRune
	keyUp
	keyDown
	keyLeft
	keyHome
	keyEnd
	keyPageUp
	keyPageDown
	keyEnter
	keyBackspace
	keyDelete
	keyTab
	keyEsc
	keyCtrlC
	keyCtrlD
	keyCtrlU
)

type keyEvent struct {
	kind keyKind
	r    rune
}

// keyReader turns the byte stream from a raw terminal into key events. A
// goroutine does the blocking read so escape sequences can be disambiguated
// with a short timeout instead of stalling on a bare Esc.
type keyReader struct {
	ch chan byte
}

func newKeyReader(f *os.File) *keyReader {
	kr := &keyReader{ch: make(chan byte, 512)}
	go func() {
		buf := make([]byte, 512)
		for {
			n, err := f.Read(buf)
			for i := 0; i < n; i++ {
				kr.ch <- buf[i]
			}
			if err != nil {
				return
			}
		}
	}()
	return kr
}

const escTimeout = 40 * time.Millisecond

func (kr *keyReader) next() (byte, bool) {
	b, ok := <-kr.ch
	return b, ok
}

func (kr *keyReader) nextWithin(d time.Duration) (byte, bool) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case b, ok := <-kr.ch:
		return b, ok
	case <-t.C:
		return 0, false
	}
}

func (kr *keyReader) read() (keyEvent, bool) {
	b, ok := kr.next()
	if !ok {
		return keyEvent{}, false
	}
	switch b {
	case 0x03:
		return keyEvent{kind: keyCtrlC}, true
	case 0x04:
		return keyEvent{kind: keyCtrlD}, true
	case 0x15:
		return keyEvent{kind: keyCtrlU}, true
	case 0x09:
		return keyEvent{kind: keyTab}, true
	case 0x0a, 0x0d:
		return keyEvent{kind: keyEnter}, true
	case 0x08, 0x7f:
		return keyEvent{kind: keyBackspace}, true
	case 0x1b:
		return kr.readEscape(), true
	}
	if b < 0x20 {
		return keyEvent{kind: keyUnknown}, true
	}
	r, ok := kr.readRune(b)
	if !ok {
		return keyEvent{}, false
	}
	return keyEvent{kind: keyRune, r: r}, true
}

func (kr *keyReader) readEscape() keyEvent {
	b, ok := kr.nextWithin(escTimeout)
	if !ok {
		return keyEvent{kind: keyEsc}
	}
	switch b {
	case '[':
		return kr.readCSI()
	case 'O':
		if f, ok := kr.nextWithin(escTimeout); ok {
			switch f {
			case 'A':
				return keyEvent{kind: keyUp}
			case 'B':
				return keyEvent{kind: keyDown}
			case 'C':
				return keyEvent{kind: keyLeft}
			case 'H':
				return keyEvent{kind: keyHome}
			case 'F':
				return keyEvent{kind: keyEnd}
			}
		}
	}
	return keyEvent{kind: keyEsc}
}

func (kr *keyReader) readCSI() keyEvent {
	params := make([]byte, 0, 8)
	for i := 0; i < 8; i++ {
		b, ok := kr.nextWithin(escTimeout)
		if !ok {
			return keyEvent{kind: keyEsc}
		}
		switch {
		case b >= '0' && b <= '9', b == ';':
			params = append(params, b)
		case b == '~':
			return csiTilde(string(params))
		default:
			switch b {
			case 'A':
				return keyEvent{kind: keyUp}
			case 'B':
				return keyEvent{kind: keyDown}
			case 'C':
				return keyEvent{kind: keyLeft}
			case 'D':
				return keyEvent{kind: keyLeft}
			case 'H':
				return keyEvent{kind: keyHome}
			case 'F':
				return keyEvent{kind: keyEnd}
			}
			return keyEvent{kind: keyUnknown}
		}
	}
	return keyEvent{kind: keyUnknown}
}

func csiTilde(p string) keyEvent {
	switch p {
	case "1", "7":
		return keyEvent{kind: keyHome}
	case "3":
		return keyEvent{kind: keyDelete}
	case "4", "8":
		return keyEvent{kind: keyEnd}
	case "5":
		return keyEvent{kind: keyPageUp}
	case "6":
		return keyEvent{kind: keyPageDown}
	}
	return keyEvent{kind: keyUnknown}
}

func (kr *keyReader) readRune(first byte) (rune, bool) {
	if first < utf8.RuneSelf {
		return rune(first), true
	}
	n := 0
	switch {
	case first&0xE0 == 0xC0:
		n = 2
	case first&0xF0 == 0xE0:
		n = 3
	case first&0xF8 == 0xF0:
		n = 4
	default:
		return utf8.RuneError, true
	}
	buf := []byte{first}
	for i := 1; i < n; i++ {
		b, ok := kr.nextWithin(escTimeout)
		if !ok {
			return utf8.RuneError, true
		}
		buf = append(buf, b)
	}
	r, _ := utf8.DecodeRune(buf)
	return r, true
}
