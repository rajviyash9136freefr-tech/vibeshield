//go:build darwin

package cli

import (
	"os"
	"syscall"
	"unsafe"
)

func ioctlTermios(fd int, req uintptr, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), req,
		uintptr(unsafe.Pointer(t)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// makeRaw switches the controlling terminal to raw mode (cf. cfmakeraw).
// macOS uses TIOCGETA / TIOCSETA rather than the Linux TCGETS / TCSETS names.
func makeRaw(f *os.File) (func(), error) {
	fd := int(f.Fd())
	var old syscall.Termios
	if err := ioctlTermios(fd, syscall.TIOCGETA, &old); err != nil {
		return nil, errNotTerminal
	}
	raw := old
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	raw.Oflag &^= syscall.OPOST
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := ioctlTermios(fd, syscall.TIOCSETA, &raw); err != nil {
		return nil, errNotTerminal
	}
	done := false
	return func() {
		if done {
			return
		}
		done = true
		_ = ioctlTermios(fd, syscall.TIOCSETA, &old)
	}, nil
}
