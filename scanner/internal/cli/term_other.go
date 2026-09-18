//go:build !windows && !linux && !darwin

package cli

import "os"

// makeRaw has no implementation on this platform, so the console degrades to
// the plain, non-interactive listing.
func makeRaw(*os.File) (func(), error) { return nil, errNotTerminal }
