//go:build windows
// +build windows

package term

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	enableEchoInput = 0x0004
	enableLineInput = 0x0002
)

// ReadPassword reads a password from the terminal without echoing it
func ReadPassword(fd int) ([]byte, error) {
	var st uint32
	r, _, err := procGetConsoleMode.Call(uintptr(fd), uintptr(unsafe.Pointer(&st)))
	if r == 0 {
		return nil, err
	}

	old := st
	st &^= (enableEchoInput | enableLineInput)
	r, _, err = procSetConsoleMode.Call(uintptr(fd), uintptr(st))
	if r == 0 {
		return nil, err
	}

	defer procSetConsoleMode.Call(uintptr(fd), uintptr(old))

	var buf [256]byte
	var n uint32
	err = syscall.ReadFile(syscall.Handle(fd), buf[:], &n, nil)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}
