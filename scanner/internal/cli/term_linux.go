//go:build linux

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
// It reports errNotTerminal when stdin is not a terminal.
func makeRaw(f *os.File) (func(), error) {
	fd := int(f.Fd())
	var old syscall.Termios
	if err := ioctlTermios(fd, syscall.TCGETS, &old); err != nil {
		return nil, errNotTerminal
	}
	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := ioctlTermios(fd, syscall.TCSETS, &raw); err != nil {
		return nil, errNotTerminal
	}
	done := false
	return func() {
		if done {
			return
		}
		done = true
		_ = ioctlTermios(fd, syscall.TCSETS, &old)
	}, nil
}
