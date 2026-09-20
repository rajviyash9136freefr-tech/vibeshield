package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/fix"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/initcmd"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scan"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/scandiff"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/tui/styles"
)

// View represents the active UI screen.
type View int

const (
	ViewHome View = iota
	ViewScanning
	ViewFindings
	ViewDetail
	ViewDiff
	ViewHelp
)

// MsgScanEvent wraps a scan event for the Bubble Tea loop.
type MsgScanEvent scan.Event

// App is the root Bubble Tea model for the VibeShield terminal UI.
type App struct {
	root        string
	version     string
	pack        *rules.Pack
	stack       initcmd.Stack
	view        View
	previous    View
	width       int
	height      int
	spinner     spinner.Model
	progress    progress.Model
	input       textinput.Model
	scanEvents  <-chan scan.Event
	cancelScan  context.CancelFunc
	currentFile string
	scanMode    string
	report      *scan.Report
	findings    []*finding.Finding
	selectedIdx int
	filterSev   string
	fixPlan     *fix.Plan
	statusMsg   string
	err         error
}

// New creates and initializes a new TUI App.
func New(root string, pack *rules.Pack, version string) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorBrand)

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	ti := textinput.New()
	ti.Placeholder = "Type a command (/scan, /diff, /fix, /help, /quit)..."
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 60

	stack, _ := initcmd.Detect(root)

	return &App{
		root:        root,
		version:     version,
		pack:        pack,
		stack:       stack,
		view:        ViewHome,
		spinner:     s,
		progress:    prog,
		input:       ti,
		findings:    []*finding.Finding{},
		scanMode:    "full",
		selectedIdx: 0,
	}
}

// Init runs startup commands.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		textinput.Blink,
	)
}

// Update processes incoming messages and user input.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.progress.Width = msg.Width - 10
		if a.progress.Width > 70 {
			a.progress.Width = 70
		}
		if a.progress.Width < 20 {
			a.progress.Width = 20
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			if a.cancelScan != nil {
				a.cancelScan()
			}
			return a, tea.Quit
		}

		// Handle keys per-view
		switch a.view {
		case ViewHome:
			return a.updateHome(msg)
		case ViewScanning:
			return a.updateScanning(msg)
		case ViewFindings:
			return a.updateFindings(msg)
		case ViewDetail:
			return a.updateDetail(msg)
		case ViewDiff:
			return a.updateDiff(msg)
		case ViewHelp:
			if msg.Type == tea.KeyEsc || msg.String() == "q" {
				a.view = a.previous
				return a, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case MsgScanEvent:
		return a.handleScanEvent(scan.Event(msg))
	}

	return a, tea.Batch(cmds...)
}

func (a *App) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		val := strings.TrimSpace(a.input.Value())
		a.input.SetValue("")
		if val == "" || val == "/scan" || val == "scan" || val == "s" {
			return a.startScan("full", "")
		}
		if val == "/diff" || val == "diff" || val == "d" {
			return a.startScan("diff", "")
		}
		if val == "/fix" || val == "fix" || val == "f" {
			if a.report != nil && len(a.report.Findings) > 0 {
				a.prepareFixPlan()
				a.view = ViewDiff
				return a, nil
			}
			a.statusMsg = "No scan findings loaded yet. Run /scan first."
			return a, nil
		}
		if val == "/help" || val == "?" || val == "help" {
			a.previous = a.view
			a.view = ViewHelp
			return a, nil
		}
		if val == "/quit" || val == "/exit" || val == "q" {
			return a, tea.Quit
		}
		a.statusMsg = fmt.Sprintf("Unknown command %q — try /scan, /diff, /fix, /help", val)
		return a, nil

	case tea.KeyEsc:
		return a, tea.Quit
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

func (a *App) updateScanning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEsc || msg.String() == "q" {
		if a.cancelScan != nil {
			a.cancelScan()
		}
		a.view = ViewHome
		a.statusMsg = "Scan cancelled."
		return a, nil
	}
	return a, nil
}

func (a *App) filteredFindings() []*finding.Finding {
	if a.filterSev == "" || a.filterSev == "all" {
		return a.findings
	}
	out := make([]*finding.Finding, 0)
	for _, f := range a.findings {
		if f.Severity == a.filterSev {
			out = append(out, f)
		}
	}
	return out
}

func (a *App) updateFindings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := a.filteredFindings()
	switch msg.String() {
	case "up", "k":
		if a.selectedIdx > 0 {
			a.selectedIdx--
		}
	case "down", "j":
		if a.selectedIdx < len(items)-1 {
			a.selectedIdx++
		}
	case "enter":
		if len(items) > 0 && a.selectedIdx < len(items) {
			a.view = ViewDetail
		}
	case "f":
		a.prepareFixPlan()
		a.view = ViewDiff
	case "s":
		return a.startScan("full", "")
	case "d":
		return a.startScan("diff", "")
	case "1":
		a.filterSev = "critical"
		a.selectedIdx = 0
	case "2":
		a.filterSev = "high"
		a.selectedIdx = 0
	case "3":
		a.filterSev = "medium"
		a.selectedIdx = 0
	case "0", "a":
		a.filterSev = "all"
		a.selectedIdx = 0
	case "esc", "q":
		a.view = ViewHome
	case "?":
		a.previous = a.view
		a.view = ViewHelp
	}
	return a, nil
}

func (a *App) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "left", "q":
		a.view = ViewFindings
	case "f":
		a.prepareFixPlan()
		a.view = ViewDiff
	case "up", "k":
		if a.selectedIdx > 0 {
			a.selectedIdx--
		}
	case "down", "j":
		items := a.filteredFindings()
		if a.selectedIdx < len(items)-1 {
			a.selectedIdx++
		}
	}
	return a, nil
}

func (a *App) updateDiff(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "n":
		a.view = ViewFindings
	case "y", "a":
		if a.fixPlan != nil && a.fixPlan.HasWork() {
			applied, err := a.applyFixes()
			if err != nil {
				a.statusMsg = fmt.Sprintf("Fix failed: %v", err)
			} else {
				a.statusMsg = fmt.Sprintf("✓ Applied %d fix(es). Run /scan to verify.", applied)
			}
			a.view = ViewHome
		}
	}
	return a, nil
}

func (a *App) prepareFixPlan() {
	if a.report == nil || a.pack == nil {
		return
	}
	a.fixPlan = fix.BuildPlan(a.root, a.report, a.pack)
}

func (a *App) applyFixes() (int, error) {
	if a.fixPlan == nil || !a.fixPlan.HasWork() {
		return 0, nil
	}
	logPath := filepath.Join(a.root, "vibeshield-fixes.log")
	audit, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer audit.Close()

	gate := fix.Gate{Mode: "yes"}
	return fix.Apply(a.root, a.fixPlan, &gate, audit, "interactive-tui")
}

func (a *App) startScan(mode, diffRef string) (tea.Model, tea.Cmd) {
	a.view = ViewScanning
	a.scanMode = mode
	a.findings = []*finding.Finding{}
	a.selectedIdx = 0
	a.statusMsg = ""

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelScan = cancel

	opts := scan.Options{}
	a.scanEvents = scan.Stream(ctx, a.root, a.pack, a.version, opts)

	return a, a.nextScanEvent()
}

func (a *App) nextScanEvent() tea.Cmd {
	return func() tea.Msg {
		if a.scanEvents == nil {
			return nil
		}
		ev, ok := <-a.scanEvents
		if !ok {
			return nil
		}
		return MsgScanEvent(ev)
	}
}

func (a *App) handleScanEvent(ev scan.Event) (tea.Model, tea.Cmd) {
	switch ev.Kind {
	case scan.EventFileScanned:
		a.currentFile = ev.CurrentFile
		return a, a.nextScanEvent()

	case scan.EventFindingDiscovered:
		if ev.Finding != nil {
			a.findings = append(a.findings, ev.Finding)
		}
		return a, a.nextScanEvent()

	case scan.EventScanComplete:
		a.report = ev.Report
		if a.scanMode == "diff" {
			d, err := scandiff.FromGit("", "--staged")
			if err == nil && d != nil {
				a.report.Scan.Mode = "diff"
				a.report.Scan.Ref = "staged"
				filterDiffFindings(a.report, d)
				a.findings = a.report.Findings
			}
		} else {
			a.findings = a.report.Findings
		}
		a.view = ViewFindings
		return a, nil

	case scan.EventScanFailed:
		a.err = ev.Error
		a.view = ViewHome
		a.statusMsg = fmt.Sprintf("Scan failed: %v", ev.Error)
		return a, nil
	}

	return a, a.nextScanEvent()
}

func filterDiffFindings(rep *scan.Report, d *scandiff.Diff) {
	kept := make([]*finding.Finding, 0)
	rep.Summary = scan.Summary{CleanFiles: rep.Summary.CleanFiles}
	for _, f := range rep.Findings {
		if d.Has(f.File, f.Line) {
			kept = append(kept, f)
			rep.Summary.Add(f.Severity)
		}
	}
	rep.Findings = kept
}

// View renders the active UI.
func (a *App) View() string {
	var b strings.Builder

	switch a.view {
	case ViewHome:
		b.WriteString(a.renderHome())
	case ViewScanning:
		b.WriteString(a.renderScanning())
	case ViewFindings:
		b.WriteString(a.renderFindings())
	case ViewDetail:
		b.WriteString(a.renderDetail())
	case ViewDiff:
		b.WriteString(a.renderDiff())
	case ViewHelp:
		b.WriteString(a.renderHelp())
	}

	return b.String()
}

func (a *App) renderHome() string {
	var b strings.Builder

	title := styles.TitleStyle.Render("🛡️  VIBESHIELD " + a.version)
	sub := styles.DimStyle.Render("AI Code Security Scanner • Local-First • Zero Telemetry")
	banner := styles.BannerBox.Render(fmt.Sprintf("%s\n%s", title, sub))
	b.WriteString("\n" + banner + "\n\n")

	// Stack & Project Info
	projName := filepath.Base(a.root)
	if projName == "." {
		if wd, err := os.Getwd(); err == nil {
			projName = filepath.Base(wd)
		}
	}
	b.WriteString(fmt.Sprintf("  %s %s\n", styles.BoldStyle.Render("Project:   "), projName))

	if !a.stack.Empty() {
		b.WriteString(fmt.Sprintf("  %s %s\n", styles.BoldStyle.Render("Stack:     "), strings.Join(a.stack.Languages, ", ")))
		if len(a.stack.Frameworks) > 0 {
			b.WriteString(fmt.Sprintf("  %s %s\n", styles.BoldStyle.Render("Frameworks:"), strings.Join(a.stack.Frameworks, ", ")))
		}
	} else {
		b.WriteString(fmt.Sprintf("  %s %s\n", styles.BoldStyle.Render("Stack:     "), styles.DimStyle.Render("Generic / Multi-language")))
	}

	if a.pack != nil {
		b.WriteString(fmt.Sprintf("  %s %d active rules loaded\n", styles.BoldStyle.Render("Engine:    "), a.pack.ActiveRules()))
	}
	b.WriteString("\n")

	// Quick Actions Card
	quickBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSubtle).
		Padding(0, 1).
		Render("Quick Keys:  [Enter / s] Full Scan    [d] Diff Scan    [f] Fix Wizard    [q] Quit")
	b.WriteString("  " + quickBox + "\n\n")

	if a.statusMsg != "" {
		b.WriteString("  " + styles.DiffPlusStyle.Render(a.statusMsg) + "\n\n")
	}

	// Prompt input
	b.WriteString("  " + styles.PromptStyle.Render("> ") + a.input.View() + "\n\n")
	b.WriteString("  " + styles.DimStyle.Render("[Tab: Autocomplete]   [?: Help]   [Ctrl+C: Quit]") + "\n")

	return b.String()
}

func (a *App) renderScanning() string {
	var b strings.Builder

	b.WriteString("\n  " + styles.TitleStyle.Render("AUDITING WORKSPACE...") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s Scanning project files: %s\n", a.spinner.View(), a.root))
	if a.currentFile != "" {
		b.WriteString("  " + styles.DimStyle.Render("Examining: "+a.currentFile) + "\n\n")
	} else {
		b.WriteString("\n")
	}

	crit, high, med, low := 0, 0, 0, 0
	for _, f := range a.findings {
		switch f.Severity {
		case "critical":
			crit++
		case "high":
			high++
		case "medium":
			med++
		case "low":
			low++
		}
	}

	b.WriteString(fmt.Sprintf("  Findings so far: %s %d Critical  %s %d High  %s %d Med  %s %d Low\n\n",
		styles.SeverityDot("critical"), crit,
		styles.SeverityDot("high"), high,
		styles.SeverityDot("medium"), med,
		styles.SeverityDot("low"), low,
	))

	b.WriteString("  " + styles.DimStyle.Render("Press [Esc] or [q] to cancel") + "\n")
	return b.String()
}

func (a *App) renderFindings() string {
	var b strings.Builder

	items := a.filteredFindings()
	total := len(a.findings)
	count := len(items)

	filterLabel := a.filterSev
	if filterLabel == "" {
		filterLabel = "all"
	}

	header := fmt.Sprintf("  %s (%d findings, filter: %s)",
		styles.TitleStyle.Render("AUDIT RESULTS"), total, filterLabel)
	b.WriteString("\n" + header + "\n")
	b.WriteString("  " + strings.Repeat("─", 78) + "\n")

	if count == 0 {
		b.WriteString("\n  " + styles.BadgeClean.Render("CLEAN") + "  No security findings at this severity!\n\n")
	} else {
		maxRows := 10
		start := 0
		if a.selectedIdx >= maxRows {
			start = a.selectedIdx - maxRows + 1
		}
		end := start + maxRows
		if end > count {
			end = count
		}

		for i := start; i < end; i++ {
			f := items[i]
			cursor := "  "
			rowStyle := lipgloss.NewStyle()
			if i == a.selectedIdx {
				cursor = "❯ "
				rowStyle = styles.SelectedRowStyle
			}

			sevDot := styles.SeverityDot(f.Severity)
			loc := fmt.Sprintf("%s:%d", f.File, f.Line)
			if len(loc) > 28 {
				loc = "..." + loc[len(loc)-25:]
			}
			title := f.Title
			if len(title) > 30 {
				title = title[:27] + "..."
			}

			line := fmt.Sprintf("%s%s %-11s %-28s %-30s",
				cursor, sevDot, f.RuleID, loc, title)
			b.WriteString(rowStyle.Render(line) + "\n")
		}
	}

	b.WriteString("  " + strings.Repeat("─", 78) + "\n")

	// Quick preview box for selected finding
	if count > 0 && a.selectedIdx < count {
		f := items[a.selectedIdx]
		previewBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBrand).
			Padding(0, 1).
			Width(76).
			Render(fmt.Sprintf("%s %s\n%s: %s\n→ Fix: %s",
				styles.SeverityBadge(f.Severity), f.Title,
				styles.DimStyle.Render(fmt.Sprintf("%s:%d", f.File, f.Line)),
				f.Snippet,
				styles.DiffPlusStyle.Render(f.Fix),
			))
		b.WriteString("  " + previewBox + "\n")
	}

	if a.statusMsg != "" {
		b.WriteString("  " + styles.DiffPlusStyle.Render(a.statusMsg) + "\n")
	}

	b.WriteString("\n  " + styles.DimStyle.Render("[↑/↓/j/k] Navigate  [Enter] Inspect  [f] Fix Wizard  [1] Crit  [2] High  [0] All  [Esc] Back") + "\n")
	return b.String()
}

func (a *App) renderDetail() string {
	var b strings.Builder

	items := a.filteredFindings()
	if len(items) == 0 || a.selectedIdx >= len(items) {
		return "No finding selected."
	}
	f := items[a.selectedIdx]

	b.WriteString("\n  " + styles.TitleStyle.Render("FINDING DETAIL: "+f.RuleID) + "\n\n")
	b.WriteString(fmt.Sprintf("  Severity: %s   Category: %s   Confidence: %.0f%%\n",
		styles.SeverityBadge(f.Severity), styles.BoldStyle.Render(f.Category), f.Confidence*100))
	b.WriteString(fmt.Sprintf("  File:     %s:%d\n\n", styles.BoldStyle.Render(f.File), f.Line))

	b.WriteString("  " + styles.BoldStyle.Render("Rule Title:") + "\n")
	b.WriteString("  " + f.Title + "\n\n")

	b.WriteString("  " + styles.BoldStyle.Render("Matched Code Snippet:") + "\n")
	snippetBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSubtle).
		Padding(0, 1).
		Render(fmt.Sprintf("%4d | %s", f.Line, f.Snippet))
	b.WriteString("  " + snippetBox + "\n\n")

	b.WriteString("  " + styles.BoldStyle.Render("Why this matters for AI code:") + "\n")
	b.WriteString("  " + f.Message + "\n\n")

	b.WriteString("  " + styles.BoldStyle.Render("Suggested Fix:") + "\n")
	b.WriteString("  " + styles.DiffPlusStyle.Render("→ "+f.Fix) + "\n\n")

	b.WriteString("  " + styles.DimStyle.Render("[f] Fix with VibePatch   [↑/↓] Previous/Next   [Esc] Back to List") + "\n")
	return b.String()
}

func (a *App) renderDiff() string {
	var b strings.Builder

	b.WriteString("\n  " + styles.TitleStyle.Render("VIBEPATCH — DIFF PREVIEW") + "\n")
	b.WriteString("  " + styles.DimStyle.Render("Mechanical, atomic line rewrites (no hallucinated changes)") + "\n\n")

	if a.fixPlan == nil || !a.fixPlan.HasWork() {
		b.WriteString("  " + styles.DimStyle.Render("Nothing to patch: no finding carries a mechanical autofix.") + "\n\n")
		b.WriteString("  " + styles.DimStyle.Render("[Esc/n] Return to findings") + "\n")
		return b.String()
	}

	for _, fp := range a.fixPlan.Files {
		b.WriteString(fmt.Sprintf("  %s (%d fixes):\n", styles.BoldStyle.Render(fp.Path), len(fp.Patches)))
		for _, pt := range fp.Patches {
			b.WriteString(fmt.Sprintf("    line %d · %s\n", pt.LineNo, pt.Fix))
			b.WriteString(fmt.Sprintf("    %s %s\n", styles.DiffMinusStyle.Render("-"), styles.DiffMinusStyle.Render(pt.Before)))
			b.WriteString(fmt.Sprintf("    %s %s\n", styles.DiffPlusStyle.Render("+"), styles.DiffPlusStyle.Render(pt.After)))
			b.WriteString("\n")
		}
	}

	b.WriteString("  Apply patches to files now?\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("[y] Apply all patches") + "    " + styles.DimStyle.Render("[n / Esc] Cancel & return") + "\n")
	return b.String()
}

func (a *App) renderHelp() string {
	var b strings.Builder

	b.WriteString("\n  " + styles.TitleStyle.Render("VIBESHIELD COMMAND GUIDE") + "\n\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("/scan [path]") + "     " + styles.HelpDescStyle.Render("Run full repository security scan") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("/diff [ref]") + "      " + styles.HelpDescStyle.Render("Audit uncommitted or git diff changes") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("/fix") + "             " + styles.HelpDescStyle.Render("Open VibePatch mechanical diff wizard") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("/help") + "            " + styles.HelpDescStyle.Render("Show this help manual") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("/quit") + "            " + styles.HelpDescStyle.Render("Exit the application") + "\n\n")

	b.WriteString("  " + styles.BoldStyle.Render("Keyboard Shortcuts:") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("↑ / ↓ / j / k") + "   " + styles.HelpDescStyle.Render("Navigate findings list and details") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("Enter") + "           " + styles.HelpDescStyle.Render("Inspect detailed finding view") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("1, 2, 3, 0") + "       " + styles.HelpDescStyle.Render("Filter by Critical, High, Medium, All") + "\n")
	b.WriteString("  " + styles.HelpKeyStyle.Render("Esc") + "             " + styles.HelpDescStyle.Render("Return to previous screen") + "\n\n")

	b.WriteString("  " + styles.DimStyle.Render("Press [Esc] or [q] to return") + "\n")
	return b.String()
}

// Run starts the Bubble Tea program.
func Run(root string, pack *rules.Pack, version string) error {
	app := New(root, pack, version)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
