package display

import (
	"errors"
	"fmt"
	"io"
	"time"

	"fyne.io/fyne/v2/dialog"

	"github.com/lucasgpulcinelli/goICMCsim/MIF"
	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
)

// fyneReadMIFCode reads the instructions from a code MIF file and loads them
// into the simulator.
func fyneReadMIFCode(f io.ReadCloser) {
	var err error

	if f == nil {
		dialog.ShowError(errors.New("reader is nil"), window)
		return
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		dialog.ShowError(err, window)
		return
	}

	data := p.GetData()
	if len(data) != 1<<16 {
		dialog.ShowError(
			fmt.Errorf("the MIF is not the right size for code: %d", len(data)),
			window,
		)
		return
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

	if err != nil {
		dialog.ShowError(err, window)
	}
}

// fyneReadMIFChar reads the character mapping definition from a MIF file and
// loads it into the simulator. The mapping can be changed while the simulator
// is running, in that case, only the next drawn characters will have the new
// char mapping, the ones that have already been drawn will stay the way they
// were.
func fyneReadMIFChar(f io.ReadCloser) {
	var err error

	if f == nil {
		dialog.ShowError(errors.New("reader is nil"), window)
		return
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		dialog.ShowError(err, window)
		return
	}

	data := p.GetData()
	if len(data) != 1<<10 {
		dialog.ShowError(
			fmt.Errorf("the MIF is not the correct size for char: %d", len(data)),
			window,
		)
		return
	}

	// set the charmap to draw with
	draw.SetCharData(p.GetData())
	draw.RedrawScreen()
	f.Close()

	if err != nil {
		dialog.ShowError(err, window)
	}
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

// getClockText returns the current processor frequency given the amount of
// instructions per second executed.
func getClockText(instructionsPerSec float32) string {
	scale := 0
	instructionsPerSec /= 1000

	for ; instructionsPerSec >= 1000 && scale < 5; scale++ {
		instructionsPerSec /= 1000
	}

	scaleStr := []string{"k", "M", "G", "T"}[scale]

	return fmt.Sprintf("clock:%8.2f %sHz", instructionsPerSec, scaleStr)
}

// updateClockLabel ticks a 100ms timer to update the clock frequency in the
// respective label. When the done channel receives a value, the function
// exits.
// this functions is expected to run in a dedicated goroutine.
func updateClockLabel(done chan struct{}) {
	instStart := icmcSimulator.InstCount

	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			instPerSec := icmcSimulator.InstCount - instStart
			periodLabel.SetText(getClockText(float32(instPerSec * 10)))
			instStart = icmcSimulator.InstCount
		}
	}
}

// runUntilHalt runs the current instruction and the next ones until a halt is
// found or the code crashes.
func runUntilHalt() {
	// do everything in a separate goroutine, because fyne uses a display
	// goroutine to run this function, meaning the display would malfunction when
	// trying to update stuff while the processor is running
	go func() {
		if icmcSimulator.IsRunning {
			dialog.ShowError(errors.New("a simulation is already running"), window)
		}

		done := make(chan struct{})
		defer func() { done <- struct{}{} }()

		go updateClockLabel(done)

		for i := 0; i < 10; i++ {
			registers[i].Disable()
		}

		simulatorMutex.Lock()
		err := icmcSimulator.RunUntilHalt(instructionPeriod)
		simulatorMutex.Unlock()

		for i := 0; i < 10; i++ {
			registers[i].Enable()
		}

		updateAllDisplay()
		if err != nil {
			dialog.ShowError(err, window)
		}
	}()
}

// runOneInst runs the instruction at the PC and increments it.
func runOneInst() {
	if icmcSimulator.IsRunning {
		dialog.ShowError(errors.New("a simulation is already running"), window)
		return
	}

	simulatorMutex.Lock()
	err := icmcSimulator.RunInstruction()
	simulatorMutex.Unlock()

	updateAllDisplay()
	if err != nil && err.Error() != "stop" {
		dialog.ShowError(err, window)
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

// toggles the visualization of instructions between raw data and operation name
func toggleInstView() {
	viewMode = viewMode * -1
	instructionList.Refresh()

}
