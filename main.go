package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/lucasgpulcinelli/goICMCsim/display"
)

var (
	initialCode = flag.String("codemif", "", "code MIF file to use at startup")
	initialChar = flag.String("charmif", "", "character MIF file to use at startup")
)

// getFiles reads from the command line flags provided both the initial code MIF
// and initial character mapping MIF, and returns them to the caller.
func getFiles() (codem, charm io.ReadCloser) {
	var err error

	// parse all the command line flags
	flag.Parse()

	if *initialCode != "" {
		codem, err = os.Open(*initialCode)
		if err != nil {
			log.Printf("error opening %s: %v\n", *initialCode, err.Error())
		}
	}
	if *initialChar != "" {
		charm, err = os.Open(*initialChar)
		if err != nil {
			log.Printf("error opening %s: %v\n", *initialChar, err.Error())
		}
	}
	return
}

func main() {
	codem, charm := getFiles()
	display.StartSimulatorWindow(codem, charm)
}
