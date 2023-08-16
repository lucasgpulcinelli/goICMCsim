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
	icmcSimulator   *processor.ICMCProcessor
	simulatorMutex  sync.Mutex
	registers       [10]*widget.Entry
	instructionList *widget.List
	currentKey      uint8 = 255
)

func makeMainMenu(w fyne.Window) *fyne.MainMenu {
	openCodeDialog := dialog.NewFileOpen(
		func(f fyne.URIReadCloser, err error) { fyneReadMIFCode(f, err) }, w)
	openCharDialog := dialog.NewFileOpen(
		func(f fyne.URIReadCloser, err error) { fyneReadMIFChar(f, err) }, w)

	file := fyne.NewMenu("file",
		fyne.NewMenuItem("open code MIF", func() { openCodeDialog.Show() }),
		fyne.NewMenuItem("open char MIF", func() { openCharDialog.Show() }),
	)

	options := fyne.NewMenu("options",
		fyne.NewMenuItem("reset", restartCode),
		fyne.NewMenuItem("run until halt", runUntilHalt),
		fyne.NewMenuItem("run one instruction", runOneInst),
	)

	help := fyne.NewMenu("help",
		fyne.NewMenuItem("show keyboard shortcuts", shortcutsHelp),
	)

	return fyne.NewMainMenu(
		file,
		options,
		help,
	)
}

func makeRegisters() fyne.CanvasObject {
	hb := container.NewGridWithColumns(1)

	for i := 0; i < 10; i++ {
		i := i

		registers[i] = widget.NewEntry()
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

		var label string

		if i < 8 {
			label = fmt.Sprintf("R%d:", i)
		} else if i == 8 {
			label = "SP:"
		} else if i == 9 {
			label = "PC:"
		}

		hb.Add(container.NewBorder(nil, nil,
			widget.NewLabel(label), nil, registers[i],
		))
	}

	return hb
}

func makeInstructionScroll() fyne.CanvasObject {
	instructionList = widget.NewList(
		func() int { return (1 << 15) - 1 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, obj fyne.CanvasObject) {
			var finalS string

			mnemonic := icmcSimulator.GetMnemonic(i)
			finalS = fmt.Sprintf("%.5d | %s", i, mnemonic)
			obj.(*widget.Label).SetText(finalS)
		},
	)

	return instructionList
}

func FyneInChar() (uint8, error) {
	ret := currentKey
	currentKey = 255
	return ret, nil
}

func StartSimulatorWindow(codem, charm io.ReadCloser) {
	icmcSimulator = processor.NewEmptyProcessor(FyneInChar, draw.FyneOutChar)

	main := app.New()
	w := main.NewWindow("ICMC Simulator")
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

	vp := draw.MakeViewPort()
	menu := makeMainMenu(w)
	regs := makeRegisters()
	insts := makeInstructionScroll()

	mainView := container.NewHSplit(
		insts,
		vp,
	)
	mainView.SetOffset(0.15)

	content := container.NewHSplit(regs, mainView)
	content.SetOffset(0.10)

	w.SetMainMenu(menu)
	w.SetContent(content)

	updateAllDisplay()
	w.Resize(fyne.NewSize(900, 500))

	if codem != nil {
		fyneReadMIFCode(codem, nil)
	}
	if charm != nil {
		fyneReadMIFChar(charm, nil)
	}

	setupShortcuts(w)
	w.ShowAndRun()
}
