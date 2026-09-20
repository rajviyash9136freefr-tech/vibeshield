package scan

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/finding"
	"github.com/rajviyash9136freefr-tech/vibeshield/scanner/internal/rules"
)

// EventKind represents the type of streaming scan event.
type EventKind int

const (
	// EventFileScanned indicates a file was examined.
	EventFileScanned EventKind = iota
	// EventFindingDiscovered indicates a new security finding was discovered.
	EventFindingDiscovered
	// EventScanComplete indicates scanning finished successfully.
	EventScanComplete
	// EventScanFailed indicates scanning was aborted by an error.
	EventScanFailed
)

// Event carries progress and findings from an asynchronous scan.
type Event struct {
	Kind         EventKind
	CurrentFile  string
	FilesScanned int
	Finding      *finding.Finding
	Report       *Report
	Error        error
}

// Stream walks root asynchronously and streams scan events over a channel.
// The channel is closed when scanning finishes or the context is cancelled.
func Stream(ctx context.Context, root string, pack *rules.Pack, version string, opts Options) <-chan Event {
	ch := make(chan Event, 64)

	go func() {
		defer close(ch)

		if pack == nil {
			ch <- Event{Kind: EventScanFailed, Error: errors.New("scan: nil rule pack")}
			return
		}

		st := time.Now()
		rep := &Report{
			SchemaVersion: 1,
			Tool:          "vibeshield",
			Version:       version,
			Scan:          ScanInfo{Mode: "full"},
			Findings:      []*finding.Finding{},
		}
		maxFiles := opts.MaxFiles
		if maxFiles <= 0 {
			maxFiles = 20000
		}

		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path != root && skipDirs[d.Name()] {
					return fs.SkipDir
				}
				return nil
			}
			if rep.Scan.FilesScanned >= maxFiles {
				return errTooManyFiles
			}
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			info, serr := d.Info()
			if serr != nil || info.Size() > MaxFileBytes {
				return nil
			}
			lang := LangForFile(path)
			if len(opts.Languages) > 0 && !langIn(opts.Languages, lang) {
				return nil
			}
			data, oerr := os.ReadFile(path)
			if oerr != nil || strings.ContainsRune(string(data), 0) {
				return nil
			}

			rep.Scan.FilesScanned++
			before := len(rep.Findings)

			scanFile(rel, string(data), lang, pack, opts, rep)

			// Notify for any new findings discovered in this file
			if len(rep.Findings) > before {
				for i := before; i < len(rep.Findings); i++ {
					select {
					case ch <- Event{
						Kind:         EventFindingDiscovered,
						CurrentFile:  rel,
						FilesScanned: rep.Scan.FilesScanned,
						Finding:      rep.Findings[i],
					}:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			} else {
				rep.Summary.CleanFiles++
			}

			// Periodic or per-file scanned tick
			select {
			case ch <- Event{
				Kind:         EventFileScanned,
				CurrentFile:  rel,
				FilesScanned: rep.Scan.FilesScanned,
			}:
			case <-ctx.Done():
				return ctx.Err()
			default:
				// non-blocking for file scanned if buffer full
			}

			return nil
		})

		if walkErr != nil && !errors.Is(walkErr, errTooManyFiles) && !errors.Is(walkErr, context.Canceled) {
			ch <- Event{Kind: EventScanFailed, Error: walkErr}
			return
		}

		rep.Scan.DurationMs = time.Since(st).Milliseconds()
		ch <- Event{
			Kind:         EventScanComplete,
			FilesScanned: rep.Scan.FilesScanned,
			Report:       rep,
		}
	}()

	return ch
}
