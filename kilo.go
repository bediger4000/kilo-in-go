package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

var origTermios *Termios

func TcSetAttr(fd uintptr, termios *Termios) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS+1), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}

func TcGetAttr(fd uintptr) (*Termios, error) {
	var termios = &Termios{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios))); err != 0 {
		return nil, err
	}
	return termios, nil
}

func enableRawMode() {
	origTermios, _ = TcGetAttr(os.Stdin.Fd())
	var raw Termios
	raw = *origTermios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	if e := TcSetAttr(os.Stdin.Fd(), &raw); e != nil {
		fmt.Fprintf(os.Stderr, "Problem enabling raw mode: %s\n", e)
	}
}

func disableRawMode() {
	if e := TcSetAttr(os.Stdin.Fd(), origTermios); e != nil {
		fmt.Fprintf(os.Stderr, "Problem disabling raw mode: %s\n", e)
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		// blank
	}
}
