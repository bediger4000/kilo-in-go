package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"
)

/*** defines ***/

const KILO_VERSION = "0.0.1"
const (
	ARROW_LEFT  = 1000 + iota
	ARROW_RIGHT = 1000 + iota
	ARROW_UP    = 1000 + iota
	ARROW_DOWN  = 1000 + iota
	DEL_KEY     = 1000 + iota
	HOME_KEY    = 1000 + iota
	END_KEY     = 1000 + iota
	PAGE_UP     = 1000 + iota
	PAGE_DOWN   = 1000 + iota
)

/*** data ***/

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

type erow struct {
	chars []byte
}

type editorConfig struct {
	cx          int
	cy          int
	rowoff      int
	coloff      int
	screenRows  int
	screenCols  int
	numRows     int
	rows        []erow
	origTermios *Termios
}

type WinSize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

var E editorConfig

/*** terminal ***/

func die(err error) {
	disableRawMode()
	io.WriteString(os.Stdout, "\x1b[2J")
	io.WriteString(os.Stdout, "\x1b[H")
	log.Fatal(err)
}

func TcSetAttr(fd uintptr, termios *Termios) error {
	// TCSETS+1 == TCSETSW, because TCSAFLUSH doesn't exist
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS+1), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}

func TcGetAttr(fd uintptr) *Termios {
	var termios = &Termios{}
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios))); err != 0 {
		log.Fatalf("Problem getting terminal attributes: %s\n", err)
	}
	return termios
}

func enableRawMode() {
	E.origTermios = TcGetAttr(os.Stdin.Fd())
	var raw Termios
	raw = *E.origTermios
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN+1] = 0
	raw.Cc[syscall.VTIME+1] = 1
	if e := TcSetAttr(os.Stdin.Fd(), &raw); e != nil {
		log.Fatalf("Problem enabling raw mode: %s\n", e)
	}
}

func disableRawMode() {
	if e := TcSetAttr(os.Stdin.Fd(), E.origTermios); e != nil {
		log.Fatalf("Problem disabling raw mode: %s\n", e)
	}
}

func editorReadKey() int {
	var buffer [1]byte
	var cc int
	var err error
	for cc, err = os.Stdin.Read(buffer[:]); cc != 1; cc, err = os.Stdin.Read(buffer[:]) {
	}
	if err != nil {
		die(err)
	}
	if buffer[0] == '\x1b' {
		var seq [2]byte
		if cc, _ = os.Stdin.Read(seq[:]); cc != 2 {
			return '\x1b'
		}

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				if cc, err = os.Stdin.Read(buffer[:]); cc != 1 {
					return '\x1b'
				}
				if buffer[0] == '~' {
					switch seq[1] {
					case '1':
						return HOME_KEY
					case '3':
						return DEL_KEY
					case '4':
						return END_KEY
					case '5':
						return PAGE_UP
					case '6':
						return PAGE_DOWN
					case '7':
						return HOME_KEY
					case '8':
						return END_KEY
					}
				}
				// XXX - what happens here?
			} else {
				switch seq[1] {
				case 'A':
					return ARROW_UP
				case 'B':
					return ARROW_DOWN
				case 'C':
					return ARROW_RIGHT
				case 'D':
					return ARROW_LEFT
				case 'H':
					return HOME_KEY
				case 'F':
					return END_KEY
				}
			}
		} else if seq[0] == '0' {
			switch seq[1] {
			case 'H':
				return HOME_KEY
			case 'F':
				return END_KEY
			}
		}

		return '\x1b'
	}
	return int(buffer[0])
}

func getCursorPosition(rows *int, cols *int) int {
	io.WriteString(os.Stdout, "\x1b[6n")
	var buffer [1]byte
	var buf []byte
	var cc int
	for cc, _ = os.Stdin.Read(buffer[:]); cc == 1; cc, _ = os.Stdin.Read(buffer[:]) {
		if buffer[0] == 'R' {
			break
		}
		buf = append(buf, buffer[0])
	}
	if string(buf[0:2]) != "\x1b[" {
		log.Printf("Failed to read rows;cols from tty\n")
		return -1
	}
	if n, e := fmt.Sscanf(string(buf[2:]), "%d;%d", rows, cols); n != 2 || e != nil {
		if e != nil {
			log.Printf("getCursorPosition: fmt.Sscanf() failed: %s\n", e)
		}
		if n != 2 {
			log.Printf("getCursorPosition: got %d items, wanted 2\n", n)
		}
		return -1
	}
	return 0
}

func getWindowSize(rows *int, cols *int) int {
	var w WinSize
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&w)),
	)
	if err != 0 { // type syscall.Errno
		io.WriteString(os.Stdout, "\x1b[999C\x1b[999B")
		return getCursorPosition(rows, cols)
	} else {
		*rows = int(w.Row)
		*cols = int(w.Col)
		return 0
	}
	return -1
}

/*** row operations ***/

func editorAppendRow(s []byte) {
	var r erow
	r.chars = s
	E.rows = append(E.rows, r)
	E.numRows++
}

/*** file I/O ***/

func editorOpen(filename string) {
	fd, err := os.Open(filename)
	if err != nil {
		die(err)
	}
	defer fd.Close()
	fp := bufio.NewReader(fd)

	for line, err := fp.ReadBytes('\n'); err == nil; line, err = fp.ReadBytes('\n') { 
		// Trim trailing newlines and carriage returns
		for c := line[len(line) - 1]; len(line) > 0 && (c == '\n' || c == '\r'); {
			line = line[:len(line)-1]
			if len(line) > 0 {
				c = line[len(line) - 1]
			}
		}
		editorAppendRow(line)
	}

	if err != nil && err != io.EOF {
		die(err)
	}
}

/*** input ***/

func editorMoveCursor(key int) {
	switch key {
	case ARROW_LEFT:
		if E.cx != 0 {
			E.cx--
		}
	case ARROW_RIGHT:
		if E.cx != E.screenCols-1 {
			E.cx++
		}
	case ARROW_UP:
		if E.cy != 0 {
			E.cy--
		}
	case ARROW_DOWN:
		if E.cy < E.numRows {
			E.cy++
		}
	}
}

func editorProcessKeypress() {
	c := editorReadKey()
	switch c {
	case ('q' & 0x1f):
		io.WriteString(os.Stdout, "\x1b[2J")
		io.WriteString(os.Stdout, "\x1b[H")
		disableRawMode()
		os.Exit(0)
	case HOME_KEY:
		E.cx = 0
	case END_KEY:
		E.cx = E.screenCols - 1
	case PAGE_UP, PAGE_DOWN:
		dir := ARROW_DOWN
		if c == PAGE_UP {
			dir = ARROW_UP
		}
		for times := E.screenRows; times > 0; times-- {
			editorMoveCursor(dir)
		}
	case ARROW_UP, ARROW_DOWN, ARROW_LEFT, ARROW_RIGHT:
		editorMoveCursor(c)
	}
}

/*** append buffer ***/

type abuf struct {
	buf []byte
}

func (p abuf) String() string {
	return string(p.buf)
}

func (p *abuf) abAppend(s string) {
	p.buf = append(p.buf, []byte(s)...)
}

func (p *abuf) abAppendBytes(b []byte) {
	p.buf = append(p.buf, b...)
}

/*** output ***/

func editorScroll() {
	if E.cy < E.rowoff {
		E.rowoff = E.cy
	}
	if E.cy >= E.rowoff + E.screenRows {
		E.rowoff = E.cy - E.screenRows + 1
	}
}

func editorRefreshScreen() {
	editorScroll()
	var ab abuf
	ab.abAppend("\x1b[25l")
	ab.abAppend("\x1b[H")
	editorDrawRows(&ab)
	ab.abAppend(fmt.Sprintf("\x1b[%d;%dH", (E.cy - E.rowoff) + 1, E.cx+1))
	ab.abAppend("\x1b[25h")
	_, e := io.WriteString(os.Stdout, ab.String())
	if e != nil {
		log.Fatal(e)
	}
}

func editorDrawRows(ab *abuf) {
	for y := 0; y < E.screenRows; y++ {
		filerow := y + E.rowoff
		if filerow >= E.numRows {
			if E.numRows == 0 && y == E.screenRows/3 {
				w := fmt.Sprintf("Kilo editor -- version %s", KILO_VERSION)
				if len(w) > E.screenCols {
					w = w[0:E.screenCols]
				}
				pad := "~ "
				for padding := (E.screenCols - len(w)) / 2; padding > 0; padding-- {
					ab.abAppend(pad)
					pad = " "
				}
				ab.abAppend(w)
			} else {
				ab.abAppend("~")
			}
		} else {
			length := len(E.rows[filerow].chars)
			if length > E.screenCols { length = E.screenCols }
			ab.abAppendBytes(E.rows[filerow].chars[:length])
		}
		ab.abAppend("\x1b[K")
		if y < E.screenRows-1 {
			ab.abAppend("\r\n")
		}
	}
}

/*** init ***/

func initEditor() {
	// Initialization a la C not necessary.
	if getWindowSize(&E.screenRows, &E.screenCols) == -1 {
		die(fmt.Errorf("couldn't get screen size"))
	}
}

func main() {
	enableRawMode()
	defer disableRawMode()
	initEditor()
	if len(os.Args) > 1 {
		editorOpen(os.Args[1])
	}

	for {
		editorRefreshScreen()
		editorProcessKeypress()
	}
}
