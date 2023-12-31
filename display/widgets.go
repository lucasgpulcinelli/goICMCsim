package display

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	registers       [10]*widget.Entry // the registers (plus SP and PC) widgets for editing
	instructionList *widget.List      // the instruction list widgets for editing
	helpPopUp       *widget.PopUp     // the popup that appears to show help
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
		fyne.NewMenuItem("stop simulation", stopSim),
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

			if icmcSimulator.IsRunning {
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

			mnemonic := icmcSimulator.GetMnemonic(i)
			finalS := fmt.Sprintf("%.5d | %s", i, mnemonic)
			obj.(*widget.Label).SetText(finalS)
		},
	)

	return instructionList
}

// makeHelpPopUp creates the popup that will appear when the user presses the
// menu button for help.
func makeHelpPopUp(w fyne.Window) {
	help := widget.NewLabel(`
  Ctrl+Tab runs a single instruction;
  Ctrl+H runs instructions until a halt, breakp or error is found;
  Ctrl+P stops execution of a simulation;
  Ctrl+O resets the simulator.
  `)
	ok := widget.NewButton("ok", func() { helpPopUp.Hide() })

	helpPopUp = widget.NewModalPopUp(
		container.NewBorder(nil, ok, nil, nil, help),
		w.Canvas(),
	)
}

// updateAllDisplay refreshes all widgets and scrolls the instruction list to
// the current instruction de PC is pointing to
func updateAllDisplay() {
	instructionList.Refresh()
	for i, reg := range registers {
		var v uint16
		switch i {
		case 8:
			v = icmcSimulator.SP
		case 9:
			v = icmcSimulator.PC
		default:
			v = icmcSimulator.GPRRegs[i]
		}
		reg.SetText(fmt.Sprintf("%d", v))
	}

	instructionList.Select(widget.ListItemID(icmcSimulator.PC))
	instructionList.ScrollTo(widget.ListItemID(icmcSimulator.PC))
}
