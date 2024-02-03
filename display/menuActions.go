package display

import (
	"io"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/lucasgpulcinelli/goICMCsim/MIF"
	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
)

// fyneReadMIFCode reads the instructions from a code MIF file and loads them
// into the simulator.
func fyneReadMIFCode(f io.ReadCloser) error {
	var err error

	if f == nil {
		return errors.New("reader is nil")
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		return err
	}

	data := p.GetData()
	if len(data) != 1<<16 {
		return fmt.Errorf("the MIF is not the right size for code: %d", len(data))
	}

	icmcSimulator.IsRunning = false
	simulatorMutex.Lock()

	// read the data in 16 bit words into the ICMC simulator code
	for i := 0; i < len(data); i += 2 {
		icmcSimulator.Code[i/2] = (uint16(data[i]) << 8) + uint16(data[i+1])
	}

	simulatorMutex.Unlock()

	// reset the viewport and restart the whole simulator, because old values for
	// registers don't make sense anymore
	draw.Reset()
	restartCode()
	f.Close()

	return err
}

// fyneReadMIFChar reads the character mapping definition from a MIF file and
// loads it into the simulator. The mapping can be changed while the simulator
// is running, in that case, only the next drawn characters will have the new
// char mapping, the ones that have already been drawn will stay the way they
// were.
func fyneReadMIFChar(f io.ReadCloser) error {
	var err error

	if f == nil {
		return errors.New("reader is nil")
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		return err
	}

	data := p.GetData()
	if len(data) != 1<<10 {
		return fmt.Errorf("the MIF is not the correct size for char: %d", len(data))
	}

	// set the charmap to draw with
	draw.SetCharData(p.GetData())
	draw.RedrawScreen()
	f.Close()

	return err
}

// restartCode resets the whole simulator to their default state,
// the same when first initialized.
func restartCode() {
	icmcSimulator.IsRunning = false

	simulatorMutex.Lock()
	icmcSimulator.Reset()
	simulatorMutex.Unlock()

	draw.Reset()
	updateAllDisplay()
}

// runUntilHalt runs the current instruction and the next ones until a halt is
// found or the code crashes.
func runUntilHalt(w fyne.Window) {
	// do everything in a separate goroutine, because fyne uses a display
	// goroutine to run this function, meaning the display would malfunction when
	// trying to update stuff while the processor is running
	
	go func(w fyne.Window) {

		if icmcSimulator.IsRunning {
			dialog.ShowError(errors.New("a simulation is already running"), w) 
			return
		}

		for i := 0; i < 10; i++ {
			registers[i].Disable()
		}

		simulatorMutex.Lock()
		err := icmcSimulator.RunUntilHalt()
		simulatorMutex.Unlock()

		for i := 0; i < 10; i++ {
			registers[i].Enable()
		}

		updateAllDisplay()
		if err != nil {
			dialog.ShowError(err, w)
		}
	}(w)
}

// runOneInst runs the instruction at the PC and increments it.
func runOneInst(w fyne.Window) {
	if icmcSimulator.IsRunning {
		dialog.ShowError(errors.New("a simulation is already running"), w)
		return
	}

	simulatorMutex.Lock()
	_, err := icmcSimulator.RunInstruction()
	simulatorMutex.Unlock()

	updateAllDisplay()
	if err != nil {
		dialog.ShowError(err, w)
	}
}

// stopSim stops the simulation if one was running
func stopSim() {
	icmcSimulator.IsRunning = false
}

// shortcutsHelp creates small help window to show shortcuts and what they do.
func shortcutsHelp() {
	helpPopUp.Show()
}
