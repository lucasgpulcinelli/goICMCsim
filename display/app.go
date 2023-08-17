// package display implements the whole user interface for the ICMC simulator.
// It is based in the fyne toolkit and is the driver for the module.
package display

import (
	"fmt"
	"io"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
	"github.com/lucasgpulcinelli/goICMCsim/processor"
)

var (
	icmcSimulator   *processor.ICMCProcessor       // the main simulator instance itself
	simulatorMutex  sync.Mutex                     // the mutex to sync simulator actions
	registers       [10]*widget.Entry              // the registers (plus SP and PC) widgets for editing
	instructionList *widget.List                   // the instruction list widgets for editing
	currentKey      uint8                    = 255 // the current key pressed by the user in ascii
)

// makeMainMenu adds in window the main menubar with all code actions
// associated. Most code actions are complex and defined in menuActions.go.
func makeMainMenu(w fyne.Window) {
	// both file dialog window popup instance (creates a little window to choose
	// a file for either a code or char MIF file)
	openCodeDialog := dialog.NewFileOpen(
		func(f fyne.URIReadCloser, err error) { fyneReadMIFCode(f, err) }, w)

	openCharDialog := dialog.NewFileOpen(
		func(f fyne.URIReadCloser, err error) { fyneReadMIFChar(f, err) }, w)

	// "file" menu toolbar
	file := fyne.NewMenu("file",
		fyne.NewMenuItem("open code MIF", func() { openCodeDialog.Show() }),
		fyne.NewMenuItem("open char MIF", func() { openCharDialog.Show() }),
	)

	// "options" menu toolbar
	options := fyne.NewMenu("options",
		fyne.NewMenuItem("reset", restartCode),
		fyne.NewMenuItem("run until halt", runUntilHalt),
		fyne.NewMenuItem("run one instruction", runOneInst),
	)

	// "help" menu toolbar
	help := fyne.NewMenu("help",
		fyne.NewMenuItem("show keyboard shortcuts", shortcutsHelp),
	)

	// with all menus, create the main one itself and associate it with the window
	w.SetMainMenu(fyne.NewMainMenu(
		file,
		options,
		help,
	))
}

// makeRegisters creates a CanvasObject with all registers (plus SP and PC)
// stacked vertically, and populates the registers global variable.
func makeRegisters() fyne.CanvasObject {
	// the stack itself
	hb := container.NewGridWithColumns(1)

	for i := 0; i < 10; i++ {
		// yes, this is necessary, this makes sure to create a new variable and not
		// just reuse the loop one, needed because of the func definition below.
		i := i

		registers[i] = widget.NewEntry()

		// define what to do when the textbox for that register is changed,
		// in our case we update the register value if it is a valid uint16
		registers[i].OnChanged = func(s string) {
			value, err := strconv.ParseUint(s, 10, 16)
			if err != nil {
				return
			}

			simulatorMutex.Lock()
			switch i {
			case 8:
				icmcSimulator.SP = uint16(value)
			case 9:
				icmcSimulator.PC = uint16(value)
			default:
				icmcSimulator.GPRRegs[i] = uint16(value)
			}
			simulatorMutex.Unlock()
		}

		// the label besides the entry textbox
		var label string
		if i < 8 {
			label = fmt.Sprintf("R%d:", i)
		} else if i == 8 {
			label = "SP:"
		} else if i == 9 {
			label = "PC:"
		}

		// add the new row for that register with a flexible center textbox,
		// and a fixed left label
		hb.Add(container.NewBorder(nil, nil,
			widget.NewLabel(label), nil, registers[i],
		))
	}

	return hb
}

// makeInstructionScroll creates a CanvasObject with a scrollable list of all
// instructions in the code loaded.
func makeInstructionScroll() fyne.CanvasObject {
	// create a new list with 2^15-1 members, empty by default, and with a certain
	// update function
	instructionList = widget.NewList(
		func() int { return (1 << 15) - 1 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, obj fyne.CanvasObject) {
			// get the mnemonic for that instruction, and display it besides it's
			// location

			var finalS string

			mnemonic := icmcSimulator.GetMnemonic(i)
			finalS = fmt.Sprintf("%.5d | %s", i, mnemonic)
			obj.(*widget.Label).SetText(finalS)
		},
	)

	return instructionList
}

// FyneInChar implements the inchar instruction for the simulator: just read
// the current key pressed and null it out to make sure the key for a single
// press is only read once by the processor.
func FyneInChar() (uint8, error) {
	ret := currentKey
	currentKey = 255
	return ret, nil
}

// setupInput creates hooks for when the user types keys while the simulator
// is running, collecting them for a possible future inchar.
func setupInput(w fyne.Window) {
	// for some reason, SetOnTypedRune does not work for enters, so SetOnTypedKey
	// is used.
	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if icmcSimulator.IsRunning {
			switch ev.Name {
			case fyne.KeyReturn:
				currentKey = '\r'
			}
		}
	})
	w.Canvas().SetOnTypedRune(func(r rune) {
		if icmcSimulator.IsRunning {
			currentKey = uint8(r)
		}
	})
}

// StartSimulatorWindow creates and starts the execution of the ICMC simulator.
// it takes as input the initial MIFs for code and character mapping.
func StartSimulatorWindow(codem, charm io.ReadCloser) {
	// create a new processor with out input and output functions
	icmcSimulator = processor.NewEmptyProcessor(FyneInChar, draw.FyneOutChar)

	// create the new fyne app, with a title and content defined in other
	// functions.
	main := app.New()
	w := main.NewWindow("ICMC Simulator")

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

	w.SetContent(content)
	makeMainMenu(w)

	w.Resize(fyne.NewSize(900, 500))

	// if the code or char mapping MIFs were defined, read them
	if codem != nil {
		fyneReadMIFCode(codem, nil)
	}
	if charm != nil {
		fyneReadMIFChar(charm, nil)
	}

	// refresh the display initially to create a proprer instruction scroll,
	// register and viewport data.
	updateAllDisplay()

	setupInput(w)
	setupShortcuts(w)

	// after everything was initialized, show the window!
	w.ShowAndRun()
}
