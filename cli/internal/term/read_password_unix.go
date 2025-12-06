//go:build unix || linux || darwin

package term

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// ReadPassword reads a line of input from the terminal without echoing the
// typed characters back to the screen.
func ReadPassword(fd int) ([]byte, error) {
	tty := os.NewFile(uintptr(fd), "/dev/tty")
	if tty == nil {
		tty = os.Stdin
	}

	disable := exec.Command("stty", "-echo")
	disable.Stdin = tty
	if err := disable.Run(); err != nil {
		return nil, err
	}
	defer func() {
		enable := exec.Command("stty", "echo")
		enable.Stdin = tty
		_ = enable.Run()
	}()

	reader := bufio.NewReader(tty)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	cleaned := strings.TrimRight(line, "\r\n")

	return []byte(cleaned), nil
}
