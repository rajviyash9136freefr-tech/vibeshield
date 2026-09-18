//go:build windows

package cli

import (
	"os"
	"syscall"
	"unsafe"
)

// Console mode bits (wincon.h).
const (
	enableProcessedInput   = 0x0001
	enableLineInput        = 0x0002
	enableEchoInput        = 0x0004
	enableWindowInput      = 0x0008
	enableVirtualTermInput = 0x0200
	enableVirtualTermProc  = 0x0004
)

// Pseudo console handles. GetStdHandle takes a DWORD, so -10 / -11 wrap.
const (
	stdInputHandle  = 0xFFFFFFF6 // (DWORD)-10
	stdOutputHandle = 0xFFFFFFF5 // (DWORD)-11
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func consoleHandle(which uintptr) (uintptr, bool) {
	h, _, _ := procGetStdHandle.Call(which)
	if h == 0 || h == ^uintptr(0) { // NULL or INVALID_HANDLE_VALUE
		return 0, false
	}
	return h, true
}

func consoleMode(h uintptr) (uint32, bool) {
	var mode uint32
	ok, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	return mode, ok != 0
}

func setConsoleMode(h uintptr, mode uint32) bool {
	ok, _, _ := procSetConsoleMode.Call(h, uintptr(mode))
	return ok != 0
}

// makeRaw puts the console into raw input + virtual-terminal mode and returns
// the undo function. It reports errNotTerminal when stdin is redirected (a
// pipe or a file), which is exactly the case the console must not block on.
//
// This is pure syscall, no CGO and no third-party dependency — the scanner
// still builds as a single static binary on every platform (CLAUDE.md rule 4).
func makeRaw(_ *os.File) (func(), error) {
	in, ok := consoleHandle(stdInputHandle)
	if !ok {
		return nil, errNotTerminal
	}
	inMode, ok := consoleMode(in)
	if !ok {
		return nil, errNotTerminal
	}
	raw := inMode &^ (enableEchoInput | enableLineInput | enableProcessedInput | enableWindowInput)
	raw |= enableVirtualTermInput
	if !setConsoleMode(in, raw) {
		return nil, errNotTerminal
	}

	// ANSI escapes on the way out need VT processing. Windows Terminal and
	// modern conhost enable it already, but cmd.exe on older builds does not.
	out, hasOut := consoleHandle(stdOutputHandle)
	var outMode uint32
	if hasOut {
		if m, ok := consoleMode(out); ok {
			outMode = m
			setConsoleMode(out, m|enableVirtualTermProc)
		}
	}

	done := false
	return func() {
		if done {
			return
		}
		done = true
		setConsoleMode(in, inMode)
		if outMode != 0 {
			setConsoleMode(out, outMode)
		}
	}, nil
}
