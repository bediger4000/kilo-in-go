package main

import (
	"os"
)

func main() {
	buffer := make([]byte, 1)
	for cc, err := os.Stdin.Read(buffer); err == nil && cc == 1; cc, err = os.Stdin.Read(buffer) {
		// blank
	}
	os.Exit(0)
}
