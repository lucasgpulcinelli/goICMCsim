package display

import (
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2/widget"
	"github.com/lucasgpulcinelli/goICMCsim/MIF"
	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
)

// updateAllDisplay refreshes the full window for the simulator and scrolls the
// instruction to the current instruction de PC is pointing to
func updateAllDisplay() {
	instructionList.Refresh()
	draw.Refresh()
	for i, widget := range registers {
		var v uint16
		switch i {
		case 8:
			v = icmcSimulator.SP
		case 9:
			v = icmcSimulator.PC
		default:
			v = icmcSimulator.GPRRegs[i]
		}
		widget.SetText(fmt.Sprintf("%d", v))
	}

	instructionList.Select(widget.ListItemID(icmcSimulator.PC))
	instructionList.ScrollTo(widget.ListItemID(icmcSimulator.PC))
}

// fyneReadMIFCode reads the instructions from a code MIF file and loads them
// into the simulator.
func fyneReadMIFCode(f io.ReadCloser, err error) {
	if err != nil {
		log.Println(err.Error())
		return
	}
	if f == nil {
		log.Println("reader is nil")
		return
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		log.Println(err.Error())
		return
	}

	data := p.GetData()
	if len(data) != 1<<16 {
		log.Printf("MIF is not the right size for code: %d\n", len(data))
		return
	}

	simulatorMutex.Lock()

	// read the data in 16 bit words into the ICMC simulator code
	for i := 0; i < len(data)/2; i += 2 {
		icmcSimulator.Code[i/2] = (uint16(data[i]) << 8) + uint16(data[i+1])
	}

	simulatorMutex.Unlock()

	// reset the viewport and restart the whole simulator, because old values for
	// registers don't make sense anymore
	draw.Reset()
	restartCode()
	f.Close()
}

// fyneReadMIFChar reads the character mapping definition from a MIF file and
// loads it into the simulator.
func fyneReadMIFChar(f io.ReadCloser, err error) {
	if err != nil {
		log.Println(err.Error())
		return
	}
	if f == nil {
		log.Println("reader is nil")
		return
	}

	// create a new MIF parser and read everything
	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		log.Println(err.Error())
		return
	}

	// set the charmap to draw with
	draw.SetCharData(p.GetData())

	// and update the display: the charmap can be changed while the simulator is
	// running!
	updateAllDisplay()
	f.Close()
}

// restartCode resets the whole simulator to their default state,
// the same when first initialized.
func restartCode() {
	simulatorMutex.Lock()
	icmcSimulator.Reset()
	simulatorMutex.Unlock()

	draw.Reset()
	updateAllDisplay()
}

// runUntilHalt runs the current instruction and the next ones until a halt is
// found or the code crashes.
func runUntilHalt() {
	// do everything in a separate goroutine, because fyne uses a display
	// goroutine to run this function, meaning the display would malfunction when
	// trying to update stuff while the processor is running
	go func() {
		simulatorMutex.Lock()
		err := icmcSimulator.RunUntilHalt()
		simulatorMutex.Unlock()

		updateAllDisplay()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}()
}

// runOneInst runs the instruction at the PC and increments it.
func runOneInst() {
	simulatorMutex.Lock()
	err := icmcSimulator.RunInstruction()
	simulatorMutex.Unlock()

	updateAllDisplay()
	if err != nil {
		log.Println(err.Error())
	}
}

// shortcutsHelp creates small help window to show shortcuts and what they do.
func shortcutsHelp() {
	updateAllDisplay()
	log.Println("not implemented")
}
