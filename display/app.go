// package display implements the whole user interface for the ICMC simulator.
// It is based in the fyne toolkit and is the driver for the module.
package display

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
	"github.com/lucasgpulcinelli/goICMCsim/processor"
)

var (
	icmcSimulator  *processor.ICMCProcessor // main simulator instance itself
	simulatorMutex sync.Mutex               // mutex to sync simulator actions

	currentKey        atomic.Value   // current key pressed by the user in ascii
	instructionPeriod *time.Duration // period between instructions
	window            fyne.Window    // main window instance for the ICMC simulator
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
	// for some reason, SetOnTypedRune does not work for some characters, so use
	// the alternative that works in these cases.
	window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if icmcSimulator.IsRunning {
			switch ev.Name {
			case fyne.KeyReturn:
				currentKey.Store(uint8('\r'))
			case fyne.KeyBackspace:
				currentKey.Store(uint8(8))
			case fyne.KeyDelete:
				currentKey.Store(uint8(127))
			case fyne.KeyEscape:
				currentKey.Store(uint8(27))
			case fyne.KeyUp:
				currentKey.Store(uint8(38))
			case fyne.KeyDown:
				currentKey.Store(uint8(40))
			case fyne.KeyLeft:
				currentKey.Store(uint8(37))
			case fyne.KeyRight:
				currentKey.Store(uint8(39))
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
	instructionPeriod = new(time.Duration)

	// initializes the first key pressed with 255
	currentKey.Store(uint8(255))

	// create a new processor with out input and output functions
	icmcSimulator = processor.NewEmptyProcessor(FyneInChar, draw.FyneOutChar)

	// create the new fyne app, with a title and content defined in other
	// functions.
	main := app.New()
	window = main.NewWindow("ICMC Simulator")

	vp := draw.MakeViewPort()
	regs := makeRegisters()
	insts := makeInstructionScroll()

	clockView := makeClockSlider()

	viewPortBorder := container.NewBorder(
		nil, clockView, nil, nil, vp,
	)

	mainView := container.NewHSplit(
		insts,
		viewPortBorder,
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
