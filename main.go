package main

import (
	"flag"
	"io"
	"os"

	"github.com/lucasgpulcinelli/goICMCsim/display"
	"github.com/sqweek/dialog"
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
			dialog.Message(
				"Reading %v failed:\n%v\n", *initialCode, err.Error()).Error()
			os.Exit(-1)
		}
	}
	if *initialChar != "" {
		charm, err = os.Open(*initialChar)
		if err != nil {
			dialog.Message(
				"Reading %v failed:\n%v\n", *initialChar, err.Error()).Error()
			os.Exit(-1)
		}
	}
	return
}

func main() {
	codem, charm := getFiles()
	display.StartSimulatorWindow(codem, charm)
}
