package display

import (
	"fmt"
	"io"
	"log"

	"fyne.io/fyne/v2/widget"
	"github.com/lucasgpulcinelli/goICMCsim/MIF"
	"github.com/lucasgpulcinelli/goICMCsim/display/draw"
)

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

func fyneReadMIFCode(f io.ReadCloser, err error) {
	if err != nil {
		log.Println(err.Error())
		return
	}
	if f == nil {
		log.Println("reader is nil")
		return
	}

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

	for i := 0; i < len(data)/2; i += 2 {
		icmcSimulator.Code[i/2] = (uint16(data[i]) << 8) + uint16(data[i+1])
	}

	simulatorMutex.Unlock()

	draw.Reset()
	restartCode()
	f.Close()
}

func fyneReadMIFChar(f io.ReadCloser, err error) {
	if err != nil {
		log.Println(err.Error())
		return
	}
	if f == nil {
		log.Println("reader is nil")
		return
	}

	p := MIF.NewParser(f)
	if err = p.Parse(); err != nil {
		log.Println(err.Error())
		return
	}

	draw.SetCharData(p.GetData())

	updateAllDisplay()
	f.Close()
}

func restartCode() {
	simulatorMutex.Lock()
	icmcSimulator.Reset()
	simulatorMutex.Unlock()

	draw.Reset()
	updateAllDisplay()
}

func runUntilHalt() {
	go func() {
		if !simulatorMutex.TryLock() {
			return
		}
		err := icmcSimulator.RunUntilHalt()
		simulatorMutex.Unlock()
		updateAllDisplay()
		if err != nil {
			log.Println(err.Error())
			return
		}
	}()
}

func runOneInst() {
	simulatorMutex.Lock()
	err := icmcSimulator.RunInstruction()
	simulatorMutex.Unlock()

	updateAllDisplay()
	if err != nil {
		log.Println(err.Error())
    icmcSimulator.PC++
	}
}

func shortcutsHelp() {
	updateAllDisplay()
	log.Println("not implemented")
}
