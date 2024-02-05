// package display implements the whole user interface for the ICMC simulator.
// It is based in the fyne toolkit and is the driver for the module.
package display

import (
	"io"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
	"github.com/lucasgpulcinelli/goICMCsim/processor"
)

var (
	icmcSimulator  *processor.ICMCProcessor // the main simulator instance itself
	simulatorMutex sync.Mutex               // the mutex to sync simulator actions

	currentKey atomic.Value // the current key pressed by the user in ascii

	window fyne.Window // the main window instance for the ICMC simulator
)

// FyneInChar implements the inchar instruction for the simulator: just read
// the current key pressed and null it out to make sure the key for a single
// press is only read once by the processor.
func FyneInChar() (uint8, error) {
	ret := currentKey.Swap(uint8(255))
	return ret.(uint8), nil
}

// setupInput creates hooks for when the user types keys while the simulator
// is running, collecting them for a possible future inchar.
func setupInput() {
	// for some reason, SetOnTypedRune does not work for enters, so SetOnTypedKey
	// is used.
	window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if icmcSimulator.IsRunning {
			switch ev.Name {
			case fyne.KeyReturn:
				currentKey.Store(uint8('\r'))
			}
		}
	})
	window.Canvas().SetOnTypedRune(func(r rune) {
		if icmcSimulator.IsRunning {
			currentKey.Store(uint8(r))
		}
	})
}

// StartSimulatorWindow creates and starts the execution of the ICMC simulator.
// it takes as input the initial MIFs for code and character mapping.
func StartSimulatorWindow(codem, charm io.ReadCloser) {
	// initializes the first key pressed with 255
	currentKey.Store(uint8(255))

	// create a new processor with out input and output functions
	icmcSimulator = processor.NewEmptyProcessor(FyneInChar, draw.FyneOutChar)

	// create the new fyne app, with a title and content defined in other
	// functions.
	main := app.New()
	w := main.NewWindow("ICMC Simulator")
	window = w

	vp := draw.MakeViewPort()
	regs := makeRegisters()
	insts := makeInstructionScroll()

	mainView := container.NewHSplit(
		insts,
		vp,
	)
	mainView.SetOffset(0.15)

	content := container.NewHSplit(regs, mainView)
	content.SetOffset(0.10)

	window.SetContent(content)
	makeMainMenu()
	makeHelpPopUp()

	window.Resize(fyne.NewSize(900, 500))

	// if the code or char mapping MIFs were defined, read them
	if codem != nil {
		fyneReadMIFCode(codem)
	}
	if charm != nil {
		fyneReadMIFChar(charm)
	}

	// refresh the display initially to create a proprer instruction scroll and
	// register data.
	updateAllDisplay()

	setupInput()
	setupShortcuts()

	// after everything was initialized, show the window!
	window.ShowAndRun()
}
