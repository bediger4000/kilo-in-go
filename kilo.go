package main

import (
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

func TcSetAttr(fd uintptr, termios *Termios) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios))); err != 0 {
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
	raw, _ := TcGetAttr(os.Stdin.Fd())
	raw.Lflag &^= syscall.ECHO
	_ = TcSetAttr(os.Stdin.Fd(), raw)
}

func main() {
	enableRawMode()
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); buffer[0] != 'q' && err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		// blank
	}
	os.Exit(0)
}
